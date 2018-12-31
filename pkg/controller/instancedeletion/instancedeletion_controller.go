/*
Copyright 2018 Cole Wippern.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package instancedeletion

import (
	"context"
	"errors"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/service/rds"
	rdsv1alpha1 "github.com/cvgw/rds-aurora-operator/pkg/apis/rds/v1alpha1"
	"github.com/cvgw/rds-aurora-operator/pkg/lib/provider"
	instanceProvider "github.com/cvgw/rds-aurora-operator/pkg/lib/provider/instance"
	kubeErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	created   = "created"
	executing = "executing"
	completed = "completed"
)

// USER ACTION REQUIRED: update cmd/manager/main.go to call this rds.Add(mgr) to install this Controller
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileInstanceDeletion{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

func add(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("instancedeletion-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &rdsv1alpha1.InstanceDeletion{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileInstanceDeletion{}

type ReconcileInstanceDeletion struct {
	client.Client
	scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=rds.nomsmon.com,resources=instances,verbs=get;list;watch
// +kubebuilder:rbac:groups=rds.nomsmon.com,resources=instancedeletions,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileInstanceDeletion) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.SetLevel(log.DebugLevel)

	logger := log.WithFields(log.Fields{
		"controller": "instance_deletion",
	})

	logger.Info("reconcile")

	result := reconcile.Result{}
	instance := &rdsv1alpha1.InstanceDeletion{}

	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if kubeErrors.IsNotFound(err) {
			logger.Debug("delete")
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	status := instance.Status
	spec := instance.Spec

	instanceList := &rdsv1alpha1.InstanceList{}
	err = r.List(context.TODO(), &client.ListOptions{}, instanceList)

	for _, i := range instanceList.Items {
		if i.Spec.Id == spec.InstanceId {
			err := errors.New("instance resource still exists, cannot delete")
			logger.Warn(err)

			return reconcile.Result{}, err
		}
	}

	sess := provider.NewSession()
	svc := rds.New(sess)

	state := status.State
	switch state {
	case "":
		logger.Debugf("setting state to %s", created)
		status.State = created
	case created:
		logger.Debugf("setting state to %s", executing)
		status.State = executing

		logger.Debug("begining deletion")
		if err := instanceProvider.DeleteDBClusterInstance(svc, spec.InstanceId); err != nil {
			logger.Warnf("error deleting instance: %s", err)
			return reconcile.Result{}, err
		}
	case executing:
		dbInstance, err := instanceProvider.FindDBClusterInstance(svc, spec.InstanceId)
		if err != nil {
			if err != instanceProvider.NotFoundErr {
				logger.Warnf("error finding instance being deleted: %s", err)
				return reconcile.Result{}, err
			}
			logger.Debugf("instance has been deleted: %s", err)
			logger.Debugf("setting state to %s", completed)
			status.State = completed
		} else {
			logger.Debug("instance is still being deleted")
			logger.Debug(dbInstance)
		}
		result.RequeueAfter = 10 * time.Second
	default:
	}

	instance.Status = status

	err = r.Status().Update(context.TODO(), instance)
	if err != nil {
		logger.Warnf("instance update failed: %s", err)
		return reconcile.Result{}, err
	}

	logger.Info("reconcile success")
	return result, nil
}

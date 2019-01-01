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

package clusterdeletion

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/service/rds"
	rdsv1alpha1 "github.com/cvgw/rds-aurora-operator/pkg/apis/rds/v1alpha1"
	"github.com/cvgw/rds-aurora-operator/pkg/lib/provider"
	"github.com/cvgw/rds-aurora-operator/pkg/lib/service"
	log "github.com/sirupsen/logrus"
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
	return &ReconcileClusterDeletion{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

func add(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("clusterdeletion-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &rdsv1alpha1.ClusterDeletion{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileClusterDeletion{}

type ReconcileClusterDeletion struct {
	client.Client
	scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=rds.nomsmon.com,resources=instances,verbs=get;list;watch
// +kubebuilder:rbac:groups=rds.nomsmon.com,resources=clusters,verbs=get;list;watch
// +kubebuilder:rbac:groups=rds.nomsmon.com,resources=clusterdeletions,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileClusterDeletion) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.SetLevel(log.DebugLevel)
	logger := log.WithFields(log.Fields{
		"controller": "cluster_deletion",
	})
	logger.Info("reconcile")

	result := reconcile.Result{}
	instance := &rdsv1alpha1.ClusterDeletion{}

	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if kubeErrors.IsNotFound(err) {
			logger.Debug("delete")
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	spec := instance.Spec
	status := instance.Status
	sess := provider.NewSession()
	svc := rds.New(sess)

	clusterList := &rdsv1alpha1.ClusterList{}
	err = r.List(context.TODO(), &client.ListOptions{}, clusterList)
	if err != nil {
		logger.Warnf("could not list clusters: %v", err)
		return reconcile.Result{RequeueAfter: 5 * time.Second}, err
	}

	for _, cluster := range clusterList.Items {
		if cluster.Spec.Id == spec.ClusterId {
			err := errors.New("cluster still exists, cannot delete")
			logger.Warn(err)

			return reconcile.Result{RequeueAfter: 5 * time.Second}, err
		}
	}

	instanceList := &rdsv1alpha1.InstanceList{}
	err = r.List(context.TODO(), &client.ListOptions{}, instanceList)
	if err != nil {
		logger.Warnf("could not list instances: %s", err)
		return reconcile.Result{}, err
	}

	for _, i := range instanceList.Items {
		if i.Spec.ClusterId == spec.ClusterId {
			err := errors.New("cluster still in use, cannot delete")
			logger.Warn(err)

			return reconcile.Result{}, err
		}
	}

	sHandler := &stateHandler{}
	sHandler.SetLogger(logger).
		SetSvc(svc).
		SetSpec(spec).
		SetStatus(&status)

	handler := &service.Handler{}
	handler.SetStateHandler(sHandler)
	result, err = service.Handle(*handler)
	if err != nil {
		return result, err
	}

	instance.Status = status
	err = r.Status().Update(context.TODO(), instance)
	if err != nil {
		logger.Warnf("instance update failed: %s", err)
		return reconcile.Result{RequeueAfter: 1 * time.Second}, err
	}

	logger.Info("reconcile success")
	return result, nil
}

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

package parametergroupdeletion

import (
	"context"
	"errors"
	"time"

	kubeErrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/aws/aws-sdk-go/service/rds"
	rdsv1alpha1 "github.com/cvgw/rds-aurora-operator/pkg/apis/rds/v1alpha1"
	"github.com/cvgw/rds-aurora-operator/pkg/lib/provider"
	"github.com/cvgw/rds-aurora-operator/pkg/lib/service"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// USER ACTION REQUIRED: update cmd/manager/main.go to call this rds.Add(mgr) to install this Controller
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileParameterGroupDeletion{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

func add(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("parametergroupdeletion-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &rdsv1alpha1.ParameterGroupDeletion{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileParameterGroupDeletion{}

type ReconcileParameterGroupDeletion struct {
	client.Client
	scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=rds.nomsmon.com,resources=parametergroupdeletions,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileParameterGroupDeletion) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.SetLevel(log.DebugLevel)
	logger := log.WithFields(log.Fields{
		"controller": "parameter_group_deletion",
	})
	logger.Info("reconcile")

	result := reconcile.Result{}
	instance := &rdsv1alpha1.ParameterGroupDeletion{}

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

	parameterGroupList := &rdsv1alpha1.ParameterGroupList{}
	err = r.List(context.TODO(), &client.ListOptions{}, parameterGroupList)
	if err != nil {
		logger.Warnf("error getting parameter groups: %s", err)
		return reconcile.Result{RequeueAfter: 5 * time.Second}, err
	}

	logger.Debugf("parameter groups: %s", parameterGroupList)
	for _, parameterGroup := range parameterGroupList.Items {
		if parameterGroup.Spec.Name == spec.ParameterGroupName {
			err := errors.New("parameter group still exists, cannot delete")
			logger.Warn(err)

			return reconcile.Result{RequeueAfter: 5 * time.Second}, err
		}
	}

	clusterList := &rdsv1alpha1.ClusterList{}
	err = r.List(context.TODO(), &client.ListOptions{}, clusterList)
	if err != nil {
		logger.Warnf("could not list clusters: %v", err)
		return reconcile.Result{RequeueAfter: 5 * time.Second}, err
	}

	for _, cluster := range clusterList.Items {
		if cluster.Spec.ParameterGroupName == spec.ParameterGroupName {
			err := errors.New("parameter group is still in use, cannot delete")
			logger.Warn(err)

			return reconcile.Result{RequeueAfter: 5 * time.Second}, err
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

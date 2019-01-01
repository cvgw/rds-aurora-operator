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

package clusterservice

import (
	"context"

	rdsv1alpha1 "github.com/cvgw/rds-aurora-operator/pkg/apis/rds/v1alpha1"
	"github.com/cvgw/rds-aurora-operator/pkg/lib/service"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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
	return &ReconcileClusterService{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

func add(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("clusterservice-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &rdsv1alpha1.ClusterService{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileClusterService{}

type ReconcileClusterService struct {
	client.Client
	scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rds.nomsmon.com,resources=clusters,verbs=get;list;watch
// +kubebuilder:rbac:groups=rds.nomsmon.com,resources=clusterservices,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileClusterService) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.SetLevel(log.DebugLevel)
	logger := log.WithFields(log.Fields{
		"controller": "cluster_service",
	})

	logger.Info("reconcile")

	result := reconcile.Result{}
	instance := &rdsv1alpha1.ClusterService{}

	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Debug("delete")

			svc := &corev1.Service{}
			svc.Name = request.NamespacedName.Name
			svc.Namespace = request.NamespacedName.Namespace

			err := r.Delete(context.TODO(), svc)
			if err != nil {
				if errors.IsNotFound(err) {
					logger.Debug("svc does not exist")
					return reconcile.Result{}, nil
				}
				logger.Warnf("error deleting svc: %s", err)
				return reconcile.Result{}, err
			}

			logger.Debug("svc deleted")
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	spec := instance.Spec
	status := instance.Status
	state := status.State

	logger.Debugf("current state is %s", state)
	switch state {
	case "":
		status.State = service.ChangeState(logger, service.Unprovisioned)
	case service.Unprovisioned:
		status.State = service.ChangeState(logger, service.Provisioning)

		cluster := &rdsv1alpha1.Cluster{}
		err := r.Get(
			context.TODO(),
			types.NamespacedName{Name: spec.ClusterName, Namespace: instance.Namespace},
			cluster,
		)
		if err != nil {
			logger.Warn(err)
			return reconcile.Result{}, err
		}

		svcName := instance.Name
		svc := &corev1.Service{}

		err = r.Get(
			context.TODO(),
			types.NamespacedName{Name: svcName, Namespace: instance.Namespace},
			svc,
		)
		if err != nil {
			if !errors.IsNotFound(err) {
				logger.Warn(err)
				return reconcile.Result{}, err
			}

			logger.Debug("svc does not exist yet")

			svc = &corev1.Service{}
			svc.Spec.Type = corev1.ServiceTypeExternalName
			svc.Spec.ExternalName = cluster.Status.Endpoint
			svc.Name = svcName
			svc.Namespace = instance.Namespace

			err := r.Create(context.TODO(), svc)
			if err != nil {
				logger.Warnf("error creating svc: %s", err)
				return reconcile.Result{}, err
			}
		} else {
			logger.Debug("svc already exists")
		}
	case service.Provisioning:
		status.State = service.ChangeState(logger, service.Provisioned)
	case service.Provisioned:
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

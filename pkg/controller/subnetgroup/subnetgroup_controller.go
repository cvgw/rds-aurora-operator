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

package subnetgroup

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/service/rds"
	rdsv1alpha1 "github.com/cvgw/rds-aurora-operator/pkg/apis/rds/v1alpha1"
	"github.com/cvgw/rds-aurora-operator/pkg/lib/provider"
	"github.com/cvgw/rds-aurora-operator/pkg/lib/service"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileSubnetGroup{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

func add(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("subnetgroup-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &rdsv1alpha1.SubnetGroup{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	//// TODO(user): Modify this to be the types you create
	//// Uncomment watch a Deployment created by SubnetGroup - change this for objects you create
	//err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
	//  IsController: true,
	//  OwnerType:    &rdsv1alpha1.SubnetGroup{},
	//})
	//if err != nil {
	//  return err
	//}

	return nil
}

var _ reconcile.Reconciler = &ReconcileSubnetGroup{}

type ReconcileSubnetGroup struct {
	client.Client
	scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=rds.nomsmon.com,resources=subnetgroups,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileSubnetGroup) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.SetLevel(log.DebugLevel)
	logger := log.WithFields(log.Fields{
		"controller": "subnet_group",
	})
	logger.Info("reconcile")

	result := reconcile.Result{}
	instance := &rdsv1alpha1.SubnetGroup{}

	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Debug("delete")
			return reconcile.Result{}, nil
		}
		logger.Warn(err)
		return reconcile.Result{}, err
	}

	spec := instance.Spec
	status := instance.Status
	sess := provider.NewSession()
	svc := rds.New(sess)

	sHandler := &stateHandler{}
	sHandler.SetLogger(logger).
		SetSvc(svc).
		SetSpec(spec).
		SetStatus(&status)

	result, err = service.HandleState(sHandler)
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

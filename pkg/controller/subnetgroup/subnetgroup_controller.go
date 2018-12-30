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

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/service/rds"
	rdsv1alpha1 "github.com/cvgw/rds-aurora-operator/pkg/apis/rds/v1alpha1"
	"github.com/cvgw/rds-aurora-operator/pkg/lib/factory/subnet_group"
	"github.com/cvgw/rds-aurora-operator/pkg/lib/provider"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	aki     = ""
	sak     = ""
	roleArn = ""
	profile = "dev"
)

func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileSubnetGroup{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("subnetgroup-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to SubnetGroup
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

// ReconcileSubnetGroup reconciles a SubnetGroup object
type ReconcileSubnetGroup struct {
	client.Client
	scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=rds.nomsmon.com,resources=subnetgroups,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileSubnetGroup) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	logger := log.WithFields(log.Fields{
		"controller": "subnet_group",
	})

	logger.Info("reconcile subnet group")

	instance := &rdsv1alpha1.SubnetGroup{}

	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("delete")
			logger.Infof("namespace named %s", request.NamespacedName)
			logger.Infof("instance name %s", instance.Name)

			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	spec := instance.Spec
	region := spec.Region

	sess := provider.NewSession(aki, sak, region, roleArn, profile)
	svc := rds.New(sess)

	group, err := subnet_group.UpdateOrCreateDBSubnetGroup(svc, spec.Name, spec.Description, spec.Subnets)
	if err != nil {
		return reconcile.Result{}, err
	}
	logger.Info(group)

	return reconcile.Result{}, nil
}

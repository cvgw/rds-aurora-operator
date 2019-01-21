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

package databaserestore

import (
	"context"

	rdsv1alpha1 "github.com/cvgw/rds-aurora-operator/pkg/apis/rds/v1alpha1"
	"github.com/cvgw/rds-aurora-operator/pkg/lib/provider/database_restore"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller")

func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileDatabaseRestore{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

func add(mgr manager.Manager, r reconcile.Reconciler) error {

	c, err := controller.New("databaserestore-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &rdsv1alpha1.DatabaseRestore{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &rdsv1alpha1.DatabaseRestore{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileDatabaseRestore{}

type ReconcileDatabaseRestore struct {
	client.Client
	scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=rds.nomsmon.com,resources=databaserestores,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rds.nomsmon.com,resources=databaserestores/status,verbs=get;update;patch
func (r *ReconcileDatabaseRestore) Reconcile(request reconcile.Request) (reconcile.Result, error) {

	instance := &rdsv1alpha1.DatabaseRestore{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {

			return reconcile.Result{}, nil
		}

		return reconcile.Result{}, err
	}

	database_restore.RestoreSQLFile(instance.Spec.Filepath)
	return reconcile.Result{}, nil
}

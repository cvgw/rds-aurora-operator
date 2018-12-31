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

package instance

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/rds"
	rdsv1alpha1 "github.com/cvgw/rds-aurora-operator/pkg/apis/rds/v1alpha1"
	"github.com/cvgw/rds-aurora-operator/pkg/lib/provider"
	instanceProvider "github.com/cvgw/rds-aurora-operator/pkg/lib/provider/instance"
	service "github.com/cvgw/rds-aurora-operator/pkg/lib/service/instance"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
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
	aki               = ""
	sak               = ""
	roleArn           = ""
	profile           = "dev"
	region            = "us-west-2"
	unprovisioned     = "unprovisioned"
	provisioning      = "provisioning"
	provisioned       = "ready"
	dbInstanceReady   = "available"
	requiredReadyTime = 120 * 1000000000
)

// USER ACTION REQUIRED: update cmd/manager/main.go to call this rds.Add(mgr) to install this Controller
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileInstance{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

func add(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("instance-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &rdsv1alpha1.Instance{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &rdsv1alpha1.Instance{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileInstance{}

type ReconcileInstance struct {
	client.Client
	scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=rds.nomsmon.com,resources=clusters,verbs=get;list;watch
// +kubebuilder:rbac:groups=rds.nomsmon.com,resources=instances,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileInstance) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.SetLevel(log.DebugLevel)
	logger := log.WithFields(log.Fields{
		"controller": "instance",
	})

	logger.Info("reconcile")

	result := reconcile.Result{}
	instance := &rdsv1alpha1.Instance{}

	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Debug("delete")
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	status := instance.Status
	spec := instance.Spec

	state := status.State

	mutateState := func(state string) {
		logger.Debugf("setting state to %s", state)
		status.State = state
	}

	sess := provider.NewSession(aki, sak, region, roleArn, profile)
	svc := rds.New(sess)

	switch state {
	case "":
		mutateState(unprovisioned)
	case unprovisioned:
		dbInstance, err := instanceProvider.FindDBClusterInstance(svc, spec.Id)
		if err != nil {
			if err != instanceProvider.NotFoundErr {
				logger.Warnf("error finding instance: %s", err)
				return reconcile.Result{}, err
			}

			logger.Debug("instance does not exist yet")

			req := service.CreateInstanceRequest{
				Spec: spec,
			}

			dbInstance, err = service.CreateDBInstance(svc, req)
			if err != nil {
				logger.Warnf("error creating instance: %s", err)
				return reconcile.Result{}, err
			}
		}

		logger.Debug(dbInstance)
		mutateState(provisioning)

		result.RequeueAfter = 10 * time.Second
	case provisioning:
		dbInstance, err := instanceProvider.FindDBClusterInstance(svc, spec.Id)
		if err != nil {
			logger.Warnf("error finding instance: %s", err)
			return reconcile.Result{}, err
		}

		log.Debug(dbInstance)

		if *dbInstance.DBInstanceStatus == dbInstanceReady {
			log.Debug("db instance is ready")

			if status.ReadySince == 0 {
				ready := time.Now().UnixNano()
				log.Debugf("setting ready since %d", ready)
				status.ReadySince = ready
			}

			readyTime := time.Now().UnixNano() - status.ReadySince
			log.Debugf("readyTime %d", readyTime)

			if readyTime >= requiredReadyTime {
				mutateState(provisioned)
			}
		} else {
			status.ReadySince = 0
		}

		result.RequeueAfter = 10 * time.Second
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

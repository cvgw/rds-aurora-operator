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

package cluster

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/rds"
	log "github.com/sirupsen/logrus"

	rdsv1alpha1 "github.com/cvgw/rds-aurora-operator/pkg/apis/rds/v1alpha1"
	"github.com/cvgw/rds-aurora-operator/pkg/lib/provider"
	clusterProvider "github.com/cvgw/rds-aurora-operator/pkg/lib/provider/cluster"
	"github.com/cvgw/rds-aurora-operator/pkg/lib/service"
	clusterService "github.com/cvgw/rds-aurora-operator/pkg/lib/service/cluster"
	"k8s.io/apimachinery/pkg/api/errors"
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
	return &ReconcileCluster{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

func add(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("cluster-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &rdsv1alpha1.Cluster{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileCluster{}

type ReconcileCluster struct {
	client.Client
	scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=rds.nomsmon.com,resources=clusters,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileCluster) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// TODO
	log.SetLevel(log.DebugLevel)

	logger := log.WithFields(log.Fields{
		"controller": "cluster",
	})

	logger.Info("reconcile")

	result := reconcile.Result{}
	instance := &rdsv1alpha1.Cluster{}

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
	state := status.State

	sess := provider.NewSession()
	svc := rds.New(sess)

	sHandler := &stateHandler{}
	sHandler.SetLogger(logger).
		SetSvc(svc).
		SetSpec(&spec).
		SetStatus(&status)

	logger.Debugf("current state is %s", state)
	switch state {
	case "":
		status.State = service.ChangeState(logger, service.Unprovisioned)
	case service.Unprovisioned:
		err := sHandler.Unprovisioned()
		if err != nil {
			return reconcile.Result{RequeueAfter: 1 * time.Second}, err
		}
	case service.Provisioning:
		err := sHandler.Provisioning()
		if err != nil {
			return reconcile.Result{RequeueAfter: 1 * time.Second}, err
		}

		result.RequeueAfter = 10 * time.Second
	case service.Provisioned:
		err := sHandler.Provisioned()
		if err != nil {
			return reconcile.Result{RequeueAfter: 1 * time.Second}, err
		}

		if status.State == service.Provisioned {
			result.RequeueAfter = 10 * time.Second
		}
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

type stateHandler struct {
	logger *log.Entry
	svc    *rds.RDS
	spec   *rdsv1alpha1.ClusterSpec
	status *rdsv1alpha1.ClusterStatus
}

func (s *stateHandler) SetLogger(v *log.Entry) *stateHandler {
	s.logger = v
	return s
}

func (s *stateHandler) SetSvc(v *rds.RDS) *stateHandler {
	s.svc = v
	return s
}

func (s *stateHandler) SetSpec(v *rdsv1alpha1.ClusterSpec) *stateHandler {
	s.spec = v
	return s
}

func (s *stateHandler) SetStatus(v *rdsv1alpha1.ClusterStatus) *stateHandler {
	s.status = v
	return s
}

func (s *stateHandler) Unprovisioned() error {
	s.status.State = service.ChangeState(s.logger, service.Provisioning)

	dbCluster, err := findOrCreateCluster(s.logger, s.svc, *s.spec)
	if err != nil {
		s.logger.Warnf("error during find or create db cluster: %s", err)
		return err
	}
	s.logger.Debug(dbCluster)

	err = clusterService.UpdateDBCluster(s.svc, dbCluster, *s.spec)
	if err != nil {
		s.logger.Warnf("error updating db cluster: %s", err)
		return err
	}

	return nil
}

func (s *stateHandler) Provisioning() error {
	dbCluster, err := clusterProvider.FindDBCluster(s.svc, s.spec.Id)
	if err != nil {
		s.logger.Warnf("error retrieving db cluster: %s", err)
		return err
	}
	s.logger.Debug(dbCluster)

	if *dbCluster.Status == service.DBClusterReady {
		s.logger.Debug("db resource is ready")

		s.status.ReadySince = service.CalculateReadySince(s.logger, s.status.ReadySince)
		s.status.State = service.StateFromReadySince(s.logger, s.status.ReadySince)
	} else {
		s.logger.Debug("db resource is not ready")
		s.status.ReadySince = 0
	}
	return nil
}

func (s *stateHandler) Provisioned() error {
	dbCluster, err := clusterProvider.FindDBCluster(s.svc, s.spec.Id)
	if err != nil {
		s.logger.Warnf("error retrieving db cluster: %s", err)
		return err
	}
	s.logger.Debug(dbCluster)

	err = clusterService.ValidateCluster(s.svc, dbCluster, *s.spec)
	if err != nil {
		s.logger.Warn(err)
		s.status.State = service.ChangeState(s.logger, service.Unprovisioned)
	} else {
		if *dbCluster.Status == service.DBClusterReady {
			s.logger.Debug("setting resource info in status")
			s.status.DBClusterId = *dbCluster.DBClusterIdentifier
			s.status.Endpoint = *dbCluster.Endpoint
			s.status.ReaderEndpoint = *dbCluster.ReaderEndpoint
		} else {
			s.logger.Debug("db resource is not ready")
			s.status.State = service.ChangeState(s.logger, service.Unprovisioned)
		}
	}

	return nil
}

func findOrCreateCluster(
	logger *log.Entry, svc *rds.RDS, spec rdsv1alpha1.ClusterSpec,
) (*rds.DBCluster, error) {

	dbCluster, err := clusterProvider.FindDBCluster(svc, spec.Id)
	if err != nil {
		if err != clusterProvider.ClusterNotFoundErr {
			logger.Warn(err)
			return nil, err
		}

		logger.Info("cluster does not exist yet")
		req := clusterService.CreateClusterRequest{
			Spec: spec,
		}
		dbCluster, err := clusterService.CreateCluster(svc, req)
		if err != nil {
			logger.Warn(err)
			return nil, err
		}

		return dbCluster, nil
	}
	logger.Info("cluster exists")

	return dbCluster, nil
}

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

const (
	aki               = ""
	sak               = ""
	roleArn           = ""
	profile           = "dev"
	region            = "us-west-2"
	unprovisioned     = "unprovisioned"
	provisioning      = "provisioning"
	provisioned       = "ready"
	dbClusterReady    = "available"
	requiredReadyTime = 120 * 1000000000
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

	instance := &rdsv1alpha1.Cluster{}

	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("delete")
			return reconcile.Result{}, nil
		}
		logger.Warn(err)
		return reconcile.Result{}, err
	}

	result := reconcile.Result{}

	copy := instance
	spec := copy.Spec
	state := copy.Status.State

	sess := provider.NewSession(aki, sak, region, roleArn, profile)
	svc := rds.New(sess)

	logger.Infof("current state is %s", state)
	switch state {
	case "":
		logger.Infof("setting state to %s", unprovisioned)
		copy.Status.State = unprovisioned
		//result.RequeueAfter = (10 * time.Second)
	case unprovisioned:
		logger.Info("setting state to %s", provisioning)
		copy.Status.State = provisioning

		dbCluster, err := findOrCreateCluster(logger, svc, spec)
		if err != nil {
			logger.Warnf("error during find or creat db cluster: %s", err)
			return reconcile.Result{}, err
		}

		logger.Info(dbCluster)
	case provisioning:
		dbCluster, err := clusterProvider.FindDBCluster(svc, spec.Id)
		if err != nil {
			logger.Warnf("error retrieving db cluster: %s", err)
			return reconcile.Result{}, err
		}

		if *dbCluster.Status == dbClusterReady {
			log.Debug("db cluster is ready")
			if copy.Status.ReadySince == 0 {
				ready := time.Now().UnixNano()
				log.Debugf("setting ready since %d", ready)
				copy.Status.ReadySince = ready
			}
		} else {
			copy.Status.ReadySince = 0
		}

		readyTime := time.Now().UnixNano() - copy.Status.ReadySince
		log.Debugf("readyTime %d", readyTime)
		if readyTime >= requiredReadyTime {
			log.Debugf("setting state to %s", provisioned)
			copy.Status.State = provisioned
		}

		result.RequeueAfter = 10 * time.Second
	default:
	}

	err = r.Status().Update(context.TODO(), copy)
	if err != nil {
		logger.Warnf("instance update failed: %s", err)
		return reconcile.Result{}, err
	}

	logger.Info("reconcile success")
	return result, nil
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

package cluster

import (
	"github.com/aws/aws-sdk-go/service/rds"
	rdsv1alpha1 "github.com/cvgw/rds-aurora-operator/pkg/apis/rds/v1alpha1"
	clusterProvider "github.com/cvgw/rds-aurora-operator/pkg/lib/provider/cluster"
	"github.com/cvgw/rds-aurora-operator/pkg/lib/service"
	clusterService "github.com/cvgw/rds-aurora-operator/pkg/lib/service/cluster"
	log "github.com/sirupsen/logrus"
)

type stateHandler struct {
	logger *log.Entry
	svc    *rds.RDS
	spec   rdsv1alpha1.ClusterSpec
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

func (s *stateHandler) SetSpec(v rdsv1alpha1.ClusterSpec) *stateHandler {
	s.spec = v
	return s
}

func (s *stateHandler) SetStatus(v *rdsv1alpha1.ClusterStatus) *stateHandler {
	s.status = v
	return s
}

func (s *stateHandler) Logger() *log.Entry {
	return s.logger
}

func (s *stateHandler) State() string {
	return s.status.State
}

func (s *stateHandler) ReadySince() int64 {
	return s.status.ReadySince
}

func (s *stateHandler) ResourceReady(resourceStatus string) bool {
	return resourceStatus == service.DBClusterReady
}

func (s *stateHandler) MutateReadySince(readySince int64) {
	s.status.ReadySince = readySince
}

func (s *stateHandler) MutateState(state string) {
	s.status.State = service.ChangeState(s.logger, state)
}

func (s *stateHandler) NoState() error {
	s.MutateState(service.Unprovisioned)
	return nil
}

func (s *stateHandler) Unprovisioned() error {
	s.MutateReadySince(0)
	s.MutateState(service.Provisioning)

	dbCluster, err := clusterProvider.FindDBCluster(s.svc, s.spec.Id)
	if err != nil {
		if err != clusterProvider.ClusterNotFoundErr {
			s.logger.Warn(err)
			return err
		}

		s.logger.Debug("cluster does not exist yet")
		req := clusterService.CreateClusterRequest{
			Spec: s.spec,
		}

		dbCluster, err = clusterService.CreateCluster(s.svc, req)
		if err != nil {
			s.logger.Warnf("error creating db cluster: %s", err)
			return err
		}
		s.logger.Debug(dbCluster)

		return nil
	}
	s.logger.Debug(dbCluster)

	if *dbCluster.Status != service.DBClusterReady {
		log.Debug("db resource is currently being modified")
		return nil
	}

	err = clusterService.UpdateDBCluster(s.svc, dbCluster, s.spec)
	if err != nil {
		s.logger.Warnf("error updating db cluster: %s", err)
		return err
	}

	log.Debug("db resource updated")
	return nil
}

func (s *stateHandler) Provisioning() error {
	dbCluster, err := clusterProvider.FindDBCluster(s.svc, s.spec.Id)
	if err != nil {
		s.logger.Warnf("error retrieving db cluster: %s", err)
		return err
	}
	s.logger.Debug(dbCluster)

	service.HandleResourceStatus(s, *dbCluster.Status)
	return nil
}

func (s *stateHandler) Provisioned() error {
	dbCluster, err := clusterProvider.FindDBCluster(s.svc, s.spec.Id)
	if err != nil {
		s.logger.Warnf("error retrieving db cluster: %s", err)
		return err
	}
	s.logger.Debug(dbCluster)

	err = clusterService.ValidateCluster(s.svc, dbCluster, s.spec)
	if err != nil {
		s.logger.Info(err)
		s.status.State = service.ChangeState(s.logger, service.Unprovisioned)
		return nil
	}

	if *dbCluster.Status == service.DBClusterReady {
		s.logger.Debug("setting resource info in status")
		s.status.DBClusterId = *dbCluster.DBClusterIdentifier
		s.status.Endpoint = *dbCluster.Endpoint
		s.status.ReaderEndpoint = *dbCluster.ReaderEndpoint
		return nil
	}

	s.logger.Debug("db resource is not ready")
	s.status.State = service.ChangeState(s.logger, service.Unprovisioned)
	return nil
}

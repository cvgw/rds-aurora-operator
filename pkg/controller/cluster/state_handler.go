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

func (s *stateHandler) NoState() error {
	s.status.State = service.ChangeState(s.logger, service.Unprovisioned)
	return nil
}

func (s *stateHandler) Unprovisioned() error {
	s.status.State = service.ChangeState(s.logger, service.Provisioning)

	dbCluster, err := findOrCreateCluster(s.logger, s.svc, s.spec)
	if err != nil {
		s.logger.Warnf("error during find or create db cluster: %s", err)
		return err
	}
	s.logger.Debug(dbCluster)

	err = clusterService.UpdateDBCluster(s.svc, dbCluster, s.spec)
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

	err = clusterService.ValidateCluster(s.svc, dbCluster, s.spec)
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

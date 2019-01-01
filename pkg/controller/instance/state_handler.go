package instance

import (
	"github.com/aws/aws-sdk-go/service/rds"
	rdsv1alpha1 "github.com/cvgw/rds-aurora-operator/pkg/apis/rds/v1alpha1"
	instanceProvider "github.com/cvgw/rds-aurora-operator/pkg/lib/provider/instance"
	"github.com/cvgw/rds-aurora-operator/pkg/lib/service"
	instanceService "github.com/cvgw/rds-aurora-operator/pkg/lib/service/instance"
	log "github.com/sirupsen/logrus"
)

type stateHandler struct {
	logger *log.Entry
	svc    *rds.RDS
	spec   rdsv1alpha1.InstanceSpec
	status *rdsv1alpha1.InstanceStatus
}

func (s *stateHandler) SetLogger(v *log.Entry) *stateHandler {
	s.logger = v
	return s
}

func (s *stateHandler) SetSvc(v *rds.RDS) *stateHandler {
	s.svc = v
	return s
}

func (s *stateHandler) SetSpec(v rdsv1alpha1.InstanceSpec) *stateHandler {
	s.spec = v
	return s
}

func (s *stateHandler) SetStatus(v *rdsv1alpha1.InstanceStatus) *stateHandler {
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
	return resourceStatus == service.DBInstanceReady
}

func (s *stateHandler) MutateReadySince(readySince int64) {
	s.status.ReadySince = readySince
}

func (s *stateHandler) MutateState(state string) {
	s.status.State = service.ChangeState(s.logger, state)
}

func (s *stateHandler) NoState() error {
	s.status.State = service.ChangeState(s.logger, service.Unprovisioned)
	return nil
}

func (s *stateHandler) Unprovisioned() error {
	s.MutateReadySince(0)
	s.MutateState(service.Provisioning)

	dbInstance, err := instanceProvider.FindDBClusterInstance(s.svc, s.spec.Id)
	if err != nil {
		if err != instanceProvider.NotFoundErr {
			s.logger.Warnf("error finding db instance: %s", err)
			return err
		}

		s.logger.Debug("db instance does not exist yet")

		req := instanceService.CreateInstanceRequest{
			Spec: s.spec,
		}

		dbInstance, err = instanceService.CreateDBInstance(s.svc, req)
		if err != nil {
			s.logger.Warnf("error creating db instance: %s", err)
			return err
		}
		s.logger.Debug(dbInstance)

		return nil
	}
	s.logger.Debug(dbInstance)

	if *dbInstance.DBInstanceStatus != service.DBInstanceReady {
		log.Debug("db resource is currently being modified")
		return nil
	}

	err = instanceService.UpdateDBInstance(s.svc, s.spec)
	if err != nil {
		s.logger.Warnf("error updating db instance: %s", err)
		return err
	}

	log.Debug("db resource updated")
	return nil
}

func (s *stateHandler) Provisioning() error {
	dbInstance, err := instanceProvider.FindDBClusterInstance(s.svc, s.spec.Id)
	if err != nil {
		s.logger.Warnf("error finding instance: %s", err)
		return err
	}
	s.logger.Debug(dbInstance)

	service.HandleResourceStatus(s, *dbInstance.DBInstanceStatus)
	return nil
}

func (s *stateHandler) Provisioned() error {
	dbInstance, err := instanceProvider.FindDBClusterInstance(s.svc, s.spec.Id)
	if err != nil {
		s.logger.Warnf("error retrieving db instance: %s", err)
		return err
	}
	s.logger.Debug(dbInstance)

	err = instanceService.ValidateInstance(s.svc, dbInstance, s.spec)
	if err != nil {
		s.logger.Info(err)
		s.status.State = service.ChangeState(s.logger, service.Unprovisioned)
		return nil
	}

	if *dbInstance.DBInstanceStatus == service.DBInstanceReady {
		s.logger.Debug("setting resource info in status")
		s.status.DBInstanceId = *dbInstance.DBInstanceIdentifier
		s.status.DBClusterId = *dbInstance.DBClusterIdentifier
		s.status.Endpoint = *dbInstance.Endpoint.Address
		return nil
	}

	s.logger.Debug("db resource is not ready")
	s.status.State = service.ChangeState(s.logger, service.Unprovisioned)
	return nil
}

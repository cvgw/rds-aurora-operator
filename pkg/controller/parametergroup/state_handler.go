package parametergroup

import (
	"github.com/aws/aws-sdk-go/service/rds"
	rdsv1alpha1 "github.com/cvgw/rds-aurora-operator/pkg/apis/rds/v1alpha1"
	paramGroupProvider "github.com/cvgw/rds-aurora-operator/pkg/lib/provider/parameter_group"
	"github.com/cvgw/rds-aurora-operator/pkg/lib/service"
	paramGroupService "github.com/cvgw/rds-aurora-operator/pkg/lib/service/parameter_group"
	log "github.com/sirupsen/logrus"
)

type stateHandler struct {
	service.BaseStateHandler
	spec   rdsv1alpha1.ParameterGroupSpec
	status *rdsv1alpha1.ParameterGroupStatus
}

func (s *stateHandler) SetLogger(v *log.Entry) *stateHandler {
	s.BaseStateHandler.Logr = v
	return s
}

func (s *stateHandler) SetSvc(v *rds.RDS) *stateHandler {
	s.BaseStateHandler.Service = v
	return s
}

func (s *stateHandler) SetSpec(v rdsv1alpha1.ParameterGroupSpec) *stateHandler {
	s.spec = v
	return s
}

func (s *stateHandler) SetStatus(v *rdsv1alpha1.ParameterGroupStatus) *stateHandler {
	s.status = v
	return s
}

func (s *stateHandler) State() string {
	return s.status.State
}

func (s *stateHandler) ReadySince() int64 {
	return s.status.ReadySince
}

func (s *stateHandler) ResourceReady(resourceStatus string) bool {
	return true
}

func (s *stateHandler) MutateReadySince(readySince int64) {
	s.status.ReadySince = readySince
}

func (s *stateHandler) MutateState(state string) {
	s.status.State = service.ChangeState(s.Logger(), state)
}

func (s *stateHandler) NoState() error {
	return nil
}

func (s *stateHandler) Unprovisioned() error {
	group, err := paramGroupProvider.FindDBParameterGroup(s.Svc(), s.spec.Name)
	if err != nil {
		if err != paramGroupProvider.NotFoundErr {
			return err
		}

		s.Logger().Debug("db parameter group does not exist yet")

		group, err = paramGroupService.CreateParameterGroup(s.Svc(), s.spec)
		if err != nil {
			s.Logger().Warnf("error creating db parameter group: %s", err)
			return err
		}
	}
	s.Logger().Debug(group)

	err = paramGroupService.UpdateParameterGroup(s.Svc(), s.spec)
	if err != nil {
		s.Logger().Warnf("error updating db parameter group: %s", err)
		return err
	}

	s.Logger().Debug("db parameter group updated")
	return nil
}

func (s *stateHandler) Provisioning() error {
	return nil
}

func (s *stateHandler) Provisioned() error {
	group, err := paramGroupProvider.FindDBParameterGroup(s.Svc(), s.spec.Name)
	if err != nil {
		s.Logger().Warnf("error retrieving db parameter group: %s", err)
		return err
	}

	err = paramGroupService.ValidateParameterGroup(s.Svc(), group, s.spec)
	if err != nil {
		s.Logger().Info(err)
		s.MutateState(service.Unprovisioned)
		return nil
	}

	return nil
}

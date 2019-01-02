package parametergroupdeletion

import (
	"github.com/aws/aws-sdk-go/service/rds"
	rdsv1alpha1 "github.com/cvgw/rds-aurora-operator/pkg/apis/rds/v1alpha1"
	parameterGroupProvider "github.com/cvgw/rds-aurora-operator/pkg/lib/provider/parameter_group"
	"github.com/cvgw/rds-aurora-operator/pkg/lib/service"
	log "github.com/sirupsen/logrus"
)

type stateHandler struct {
	service.BaseStateHandler
	spec   rdsv1alpha1.ParameterGroupDeletionSpec
	status *rdsv1alpha1.ParameterGroupDeletionStatus
}

func (s *stateHandler) SetLogger(v *log.Entry) *stateHandler {
	s.BaseStateHandler.Logr = v
	return s
}

func (s *stateHandler) SetSvc(v *rds.RDS) *stateHandler {
	s.BaseStateHandler.Service = v
	return s
}

func (s *stateHandler) SetSpec(v rdsv1alpha1.ParameterGroupDeletionSpec) *stateHandler {
	s.spec = v
	return s
}

func (s *stateHandler) SetStatus(v *rdsv1alpha1.ParameterGroupDeletionStatus) *stateHandler {
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
	if err := parameterGroupProvider.DeleteDBParameterGroup(s.Svc(), s.spec.ParameterGroupName); err != nil {
		s.Logger().Warnf("error deleting parameterGroup: %s", err)
		return err
	}
	s.Logger().Debug("db parameterGroup deletion request successful")
	return nil
}

func (s *stateHandler) Provisioning() error {
	dbParameterGroup, err := parameterGroupProvider.FindDBParameterGroup(s.Svc(), s.spec.ParameterGroupName)
	if err != nil {
		if err != parameterGroupProvider.NotFoundErr {
			s.Logger().Warnf("error finding parameter group being deleted: %s", err)
			return err
		}
		s.Logger().Debugf("parameter group has been deleted: %s", err)
		s.MutateState(service.Provisioned)
	} else {
		s.Logger().Debug("parameter group is still being deleted")
		s.Logger().Debug(dbParameterGroup)
	}
	return nil
}

func (s *stateHandler) Provisioned() error {
	return nil
}

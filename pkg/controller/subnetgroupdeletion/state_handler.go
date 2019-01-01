package subnetgroupdeletion

import (
	"github.com/aws/aws-sdk-go/service/rds"
	rdsv1alpha1 "github.com/cvgw/rds-aurora-operator/pkg/apis/rds/v1alpha1"
	subnetGroupProvider "github.com/cvgw/rds-aurora-operator/pkg/lib/provider/subnet_group"
	"github.com/cvgw/rds-aurora-operator/pkg/lib/service"
	log "github.com/sirupsen/logrus"
)

type stateHandler struct {
	service.BaseStateHandler
	spec   rdsv1alpha1.SubnetGroupDeletionSpec
	status *rdsv1alpha1.SubnetGroupDeletionStatus
}

func (s *stateHandler) SetLogger(v *log.Entry) *stateHandler {
	s.BaseStateHandler.Logr = v
	return s
}

func (s *stateHandler) SetSvc(v *rds.RDS) *stateHandler {
	s.BaseStateHandler.Service = v
	return s
}

func (s *stateHandler) SetSpec(v rdsv1alpha1.SubnetGroupDeletionSpec) *stateHandler {
	s.spec = v
	return s
}

func (s *stateHandler) SetStatus(v *rdsv1alpha1.SubnetGroupDeletionStatus) *stateHandler {
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
	if err := subnetGroupProvider.DeleteDBSubnetGroup(s.Svc(), s.spec.SubnetGroupName); err != nil {
		s.Logger().Warnf("error deleting subnetGroup: %s", err)
		return err
	}
	s.Logger().Debug("db subnetGroup deletion request successful")
	return nil
}

func (s *stateHandler) Provisioning() error {
	dbSubnetGroup, err := subnetGroupProvider.FindDBSubnetGroup(s.Svc(), s.spec.SubnetGroupName)
	if err != nil {
		if err != subnetGroupProvider.SubnetGroupNotFoundErr {
			s.Logger().Warnf("error finding subnet group being deleted: %s", err)
			return err
		}
		s.Logger().Debugf("subnet group has been deleted: %s", err)
		s.MutateState(service.Provisioned)
	} else {
		s.Logger().Debug("subnet group is still being deleted")
		s.Logger().Debug(dbSubnetGroup)
	}
	return nil
}

func (s *stateHandler) Provisioned() error {
	return nil
}

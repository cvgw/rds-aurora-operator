package subnetgroup

import (
	"github.com/aws/aws-sdk-go/service/rds"
	rdsv1alpha1 "github.com/cvgw/rds-aurora-operator/pkg/apis/rds/v1alpha1"
	"github.com/cvgw/rds-aurora-operator/pkg/lib/factory/subnet_group"
	"github.com/cvgw/rds-aurora-operator/pkg/lib/service"
	log "github.com/sirupsen/logrus"
)

type stateHandler struct {
	logger *log.Entry
	svc    *rds.RDS
	spec   rdsv1alpha1.SubnetGroupSpec
	status *rdsv1alpha1.SubnetGroupStatus
}

func (s *stateHandler) SetLogger(v *log.Entry) *stateHandler {
	s.logger = v
	return s
}

func (s *stateHandler) SetSvc(v *rds.RDS) *stateHandler {
	s.svc = v
	return s
}

func (s *stateHandler) SetSpec(v rdsv1alpha1.SubnetGroupSpec) *stateHandler {
	s.spec = v
	return s
}

func (s *stateHandler) SetStatus(v *rdsv1alpha1.SubnetGroupStatus) *stateHandler {
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

func (s *stateHandler) ResourceReady(_ string) bool {
	return true
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

	group, err := subnet_group.UpdateOrCreateDBSubnetGroup(
		s.svc,
		s.spec.Name,
		s.spec.Description,
		s.spec.Subnets,
	)
	if err != nil {
		return err
	}
	s.Logger().Debug(group)

	s.Logger().Debug("db resource updated")
	return nil
}

func (s *stateHandler) Provisioning() error {
	service.HandleResourceStatus(s, "")
	return nil
}

func (s *stateHandler) Provisioned() error {
	return nil
}

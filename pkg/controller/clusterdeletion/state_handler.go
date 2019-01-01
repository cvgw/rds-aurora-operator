package clusterdeletion

import (
	"github.com/aws/aws-sdk-go/service/rds"
	rdsv1alpha1 "github.com/cvgw/rds-aurora-operator/pkg/apis/rds/v1alpha1"
	clusterProvider "github.com/cvgw/rds-aurora-operator/pkg/lib/provider/cluster"
	"github.com/cvgw/rds-aurora-operator/pkg/lib/service"
	log "github.com/sirupsen/logrus"
)

type stateHandler struct {
	service.BaseStateHandler
	spec   rdsv1alpha1.ClusterDeletionSpec
	status *rdsv1alpha1.ClusterDeletionStatus
}

func (s *stateHandler) SetLogger(v *log.Entry) *stateHandler {
	s.BaseStateHandler.Logr = v
	return s
}

func (s *stateHandler) SetSvc(v *rds.RDS) *stateHandler {
	s.BaseStateHandler.Service = v
	return s
}

func (s *stateHandler) SetSpec(v rdsv1alpha1.ClusterDeletionSpec) *stateHandler {
	s.spec = v
	return s
}

func (s *stateHandler) SetStatus(v *rdsv1alpha1.ClusterDeletionStatus) *stateHandler {
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
	if err := clusterProvider.DeleteDBCluster(s.Svc(), s.spec.ClusterId); err != nil {
		s.Logger().Warnf("error deleting cluster: %s", err)
		return err
	}
	s.Logger().Debug("db cluster deletion request successful")
	return nil
}

func (s *stateHandler) Provisioning() error {
	dbCluster, err := clusterProvider.FindDBCluster(s.Svc(), s.spec.ClusterId)
	if err != nil {
		if err != clusterProvider.ClusterNotFoundErr {
			s.Logger().Warnf("error finding cluster being deleted: %s", err)
			return err
		}
		s.Logger().Debugf("cluster has been deleted: %s", err)
		s.MutateState(service.Provisioned)
	} else {
		s.Logger().Debug("cluster is still being deleted")
		s.Logger().Debug(dbCluster)
	}
	return nil
}

func (s *stateHandler) Provisioned() error {
	return nil
}

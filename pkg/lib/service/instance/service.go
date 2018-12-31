package instance

import (
	"github.com/aws/aws-sdk-go/service/rds"
	rdsv1alpha1 "github.com/cvgw/rds-aurora-operator/pkg/apis/rds/v1alpha1"
	instanceFactory "github.com/cvgw/rds-aurora-operator/pkg/lib/factory/instance"
)

type CreateInstanceRequest struct {
	Spec rdsv1alpha1.InstanceSpec
}

func CreateDBInstance(svc *rds.RDS, req CreateInstanceRequest) (*rds.DBInstance, error) {
	f := instanceFactory.DBInstanceFactory{}
	f.SetSvc(svc)
	f.SetInstanceIdentifier(req.Spec.Id)
	f.SetClusterIdentifier(req.Spec.ClusterId)
	f.SetAllocatedStorage(req.Spec.AllocatedStorage)
	f.SetEngine(req.Spec.Engine)
	f.SetInstanceClass(req.Spec.Class)

	return f.CreateDBClusterInstance()
}

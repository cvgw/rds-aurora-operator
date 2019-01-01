package instance

import (
	"github.com/aws/aws-sdk-go/service/rds"
	rdsv1alpha1 "github.com/cvgw/rds-aurora-operator/pkg/apis/rds/v1alpha1"
	instanceFactory "github.com/cvgw/rds-aurora-operator/pkg/lib/factory/instance"
	instanceProvider "github.com/cvgw/rds-aurora-operator/pkg/lib/provider/instance"
)

type CreateInstanceRequest struct {
	Spec rdsv1alpha1.InstanceSpec
}

func CreateDBInstance(svc *rds.RDS, req CreateInstanceRequest) (*rds.DBInstance, error) {
	f := instanceFactory.DBInstanceFactory{}
	f.SetSvc(svc).
		SetInstanceIdentifier(req.Spec.Id).
		SetClusterIdentifier(req.Spec.ClusterId).
		SetAllocatedStorage(req.Spec.AllocatedStorage).
		SetEngine(req.Spec.Engine).
		SetInstanceClass(req.Spec.Class)

	return f.CreateDBClusterInstance()
}

func UpdateDBInstance(svc *rds.RDS, spec rdsv1alpha1.InstanceSpec) error {
	req := &instanceProvider.UpdateDBInstanceRequest{}
	req.SetId(spec.Id).
		SetClusterId(spec.ClusterId).
		SetAllocatedStorage(spec.AllocatedStorage).
		SetEngine(spec.Engine).
		SetClass(spec.Class)

	return instanceProvider.UpdateDBClusterInstance(svc, *req)
}

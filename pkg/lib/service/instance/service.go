package instance

import (
	"errors"

	"github.com/aws/aws-sdk-go/service/rds"
	rdsv1alpha1 "github.com/cvgw/rds-aurora-operator/pkg/apis/rds/v1alpha1"
	instanceFactory "github.com/cvgw/rds-aurora-operator/pkg/lib/factory/instance"
	instanceProvider "github.com/cvgw/rds-aurora-operator/pkg/lib/provider/instance"
	"github.com/cvgw/rds-aurora-operator/pkg/lib/service"
)

type CreateInstanceRequest struct {
	Spec rdsv1alpha1.InstanceSpec
}

func CreateDBInstance(svc *rds.RDS, req CreateInstanceRequest) (*rds.DBInstance, error) {
	f := instanceFactory.DBInstanceFactory{}
	f.SetSvc(svc).
		SetInstanceIdentifier(req.Spec.Id).
		SetClusterIdentifier(req.Spec.ClusterId).
		SetEngine(req.Spec.Engine).
		SetAllocatedStorage(req.Spec.AllocatedStorage).
		SetInstanceClass(req.Spec.Class)

	return f.CreateDBClusterInstance()
}

func UpdateDBInstance(svc *rds.RDS, spec rdsv1alpha1.InstanceSpec) error {
	req := &instanceProvider.UpdateDBInstanceRequest{}
	req.SetId(spec.Id).
		SetClusterId(spec.ClusterId).
		SetEngine(spec.Engine).
		SetAllocatedStorage(spec.AllocatedStorage).
		SetClass(spec.Class)

	return instanceProvider.UpdateDBClusterInstance(svc, *req)
}

func ValidateInstance(svc *rds.RDS, dbInstance *rds.DBInstance, spec rdsv1alpha1.InstanceSpec) error {
	var err error

	if *dbInstance.DBInstanceIdentifier != spec.Id {
		err = service.PopulateValidationErr(err, errors.New("db instance identifier does not match"))
	}

	if *dbInstance.Engine != spec.Engine {
		err = service.PopulateValidationErr(err, errors.New("db engine does not match"))
	}

	//if *dbInstance.EngineVersion != spec.EngineVersion {
	//  err = service.PopulateValidationErr(err, errors.New("db engine version does not match"))
	//}

	if *dbInstance.DBInstanceClass != spec.Class {
		err = service.PopulateValidationErr(err, errors.New("db instance class does not match"))
	}

	return err
}

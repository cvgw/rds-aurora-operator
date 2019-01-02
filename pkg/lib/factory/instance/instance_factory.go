package factory

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	log "github.com/sirupsen/logrus"
)

var (
	notFoundErr error
)

func init() {
	notFoundErr = errors.New("not found")
}

type DBInstanceFactory struct {
	svc                *rds.RDS
	instanceIdentifier *string
	clusterIdentifier  *string
	allocatedStorage   *int64
	engine             *string
	instanceClass      *string
	parameterGroupName *string
}

func (f *DBInstanceFactory) SetSvc(v *rds.RDS) *DBInstanceFactory {
	f.svc = v
	return f
}

func (f *DBInstanceFactory) SetInstanceIdentifier(v string) *DBInstanceFactory {
	f.instanceIdentifier = aws.String(v)
	return f
}

func (f *DBInstanceFactory) SetClusterIdentifier(v string) *DBInstanceFactory {
	f.clusterIdentifier = aws.String(v)
	return f
}

func (f *DBInstanceFactory) SetAllocatedStorage(v int) *DBInstanceFactory {
	f.allocatedStorage = aws.Int64(int64(v))
	return f
}

func (f *DBInstanceFactory) SetEngine(v string) *DBInstanceFactory {
	f.engine = aws.String(v)
	return f
}

func (f *DBInstanceFactory) SetInstanceClass(v string) *DBInstanceFactory {
	f.instanceClass = aws.String(v)
	return f
}

func (f *DBInstanceFactory) SetParameterGroupName(v string) *DBInstanceFactory {
	//names := make([]*string, len(v))
	//for i, name := range v {
	//  names[i] = aws.String(name)
	//}
	//f.parameterGroupNames = names
	f.parameterGroupName = aws.String(v)
	return f
}

func (f *DBInstanceFactory) CreateDBClusterInstance() (*rds.DBInstance, error) {

	instanceInput := &rds.CreateDBInstanceInput{
		DBInstanceIdentifier: f.instanceIdentifier,
		DBClusterIdentifier:  f.clusterIdentifier,
		Engine:               f.engine,
		DBInstanceClass:      f.instanceClass,
		DBParameterGroupName: f.parameterGroupName,
	}

	instanceOutput, err := f.svc.CreateDBInstance(instanceInput)
	if err != nil {
		log.Warn(err)
		return nil, err
	}

	return instanceOutput.DBInstance, nil
}

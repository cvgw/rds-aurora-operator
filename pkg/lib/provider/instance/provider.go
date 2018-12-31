package instance

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"
	log "github.com/sirupsen/logrus"
)

var (
	NotFoundErr error
)

func init() {
	NotFoundErr = errors.New("db instance not found")
}

func FindDBClusterInstance(svc *rds.RDS, instanceId string) (*rds.DBInstance, error) {
	descInstancesInput := &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(instanceId),
	}

	descInstancesOuput, err := svc.DescribeDBInstances(descInstancesInput)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case rds.ErrCodeDBInstanceNotFoundFault:
				log.Info(rds.ErrCodeDBInstanceNotFoundFault, aerr.Error())
				return nil, NotFoundErr
			default:
				log.Warn(aerr.Error())
				return nil, aerr
			}
		} else {
			log.Warn(err.Error())
			return nil, err
		}
	}

	return descInstancesOuput.DBInstances[0], nil
}

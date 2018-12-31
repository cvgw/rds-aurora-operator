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

func DeleteDBClusterInstance(svc *rds.RDS, instanceId string) error {
	input := &rds.DeleteDBInstanceInput{
		DBInstanceIdentifier: aws.String(instanceId),
		SkipFinalSnapshot:    aws.Bool(true),
	}

	_, err := svc.DeleteDBInstance(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case rds.ErrCodeDBInstanceNotFoundFault:
				log.Warn(rds.ErrCodeDBInstanceNotFoundFault, aerr.Error())
				return NotFoundErr
			case rds.ErrCodeInvalidDBInstanceStateFault:
				log.Warn(rds.ErrCodeInvalidDBInstanceStateFault, aerr.Error())
				return aerr
			case rds.ErrCodeDBSnapshotAlreadyExistsFault:
				log.Warn(rds.ErrCodeDBSnapshotAlreadyExistsFault, aerr.Error())
				return aerr
			case rds.ErrCodeSnapshotQuotaExceededFault:
				log.Warn(rds.ErrCodeSnapshotQuotaExceededFault, aerr.Error())
				return aerr
			case rds.ErrCodeInvalidDBClusterStateFault:
				log.Warn(rds.ErrCodeInvalidDBClusterStateFault, aerr.Error())
				return aerr
			case rds.ErrCodeDBInstanceAutomatedBackupQuotaExceededFault:
				log.Warn(rds.ErrCodeDBInstanceAutomatedBackupQuotaExceededFault, aerr.Error())
				return aerr
			default:
				log.Warn(aerr.Error())
				return aerr
			}
		} else {
			log.Warn(err.Error())
			return err
		}
	}

	return nil
}

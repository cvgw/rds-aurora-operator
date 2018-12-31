package cluster

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"
	log "github.com/sirupsen/logrus"
)

func DeleteDBCluster(svc *rds.RDS, clusterId string) error {
	input := &rds.DeleteDBClusterInput{
		DBClusterIdentifier: aws.String(clusterId),
		SkipFinalSnapshot:   aws.Bool(true),
	}

	_, err := svc.DeleteDBCluster(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case rds.ErrCodeDBClusterNotFoundFault:
				log.Warn(rds.ErrCodeDBClusterNotFoundFault, aerr.Error())
				return aerr
			case rds.ErrCodeInvalidDBClusterStateFault:
				log.Warn(rds.ErrCodeInvalidDBClusterStateFault, aerr.Error())
				return aerr
			case rds.ErrCodeDBClusterSnapshotAlreadyExistsFault:
				log.Warn(rds.ErrCodeDBClusterSnapshotAlreadyExistsFault, aerr.Error())
				return aerr
			case rds.ErrCodeSnapshotQuotaExceededFault:
				log.Warn(rds.ErrCodeSnapshotQuotaExceededFault, aerr.Error())
				return aerr
			case rds.ErrCodeInvalidDBClusterSnapshotStateFault:
				log.Warn(rds.ErrCodeInvalidDBClusterSnapshotStateFault, aerr.Error())
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

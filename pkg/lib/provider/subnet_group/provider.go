package subnet_group

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"
	log "github.com/sirupsen/logrus"
)

var (
	SubnetGroupNotFoundErr error
)

func init() {
	SubnetGroupNotFoundErr = errors.New("subnet group not found")
}

func FindDBSubnetGroup(svc *rds.RDS, groupName string) (*rds.DBSubnetGroup, error) {
	subnetGroupName := aws.String(groupName)
	descGroupsInput := &rds.DescribeDBSubnetGroupsInput{
		DBSubnetGroupName: subnetGroupName,
	}

	descGroupsOutput, err := svc.DescribeDBSubnetGroups(descGroupsInput)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case rds.ErrCodeDBSubnetGroupNotFoundFault:
				log.Debug(rds.ErrCodeDBSubnetGroupNotFoundFault, aerr.Error())
				return nil, SubnetGroupNotFoundErr
			default:
				log.Warn(aerr)
				return nil, aerr
			}
		} else {
			log.Warn(err)
			return nil, err
		}
	}

	return descGroupsOutput.DBSubnetGroups[0], nil
}

func DeleteDBSubnetGroup(svc *rds.RDS, groupName string) error {
	input := &rds.DeleteDBSubnetGroupInput{
		DBSubnetGroupName: aws.String(groupName),
	}

	result, err := svc.DeleteDBSubnetGroup(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case rds.ErrCodeInvalidDBSubnetGroupStateFault:
				log.Warn(rds.ErrCodeInvalidDBSubnetGroupStateFault, aerr.Error())
				return aerr
			case rds.ErrCodeInvalidDBSubnetStateFault:
				log.Warn(rds.ErrCodeInvalidDBSubnetStateFault, aerr.Error())
				return aerr
			case rds.ErrCodeDBSubnetGroupNotFoundFault:
				log.Warn(rds.ErrCodeDBSubnetGroupNotFoundFault, aerr.Error())
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
	log.Debug(result)

	return nil
}

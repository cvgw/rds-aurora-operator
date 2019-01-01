package subnet_group

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"
	subnetGroupProvider "github.com/cvgw/rds-aurora-operator/pkg/lib/provider/subnet_group"
	log "github.com/sirupsen/logrus"
)

func UpdateOrCreateDBSubnetGroup(svc *rds.RDS, groupName, groupDescription string, subnets []string) (*rds.DBSubnetGroup, error) {
	var subnetGroup *rds.DBSubnetGroup

	subnetGroup, err := subnetGroupProvider.FindDBSubnetGroup(svc, groupName)
	if err != nil {
		if err != subnetGroupProvider.SubnetGroupNotFoundErr {
			return nil, err
		}
		subnetGroup, err = createSubnetGroup(svc, aws.String(groupName), groupDescription, subnets)
		if err != nil {
			return nil, err
		}
	}

	return subnetGroup, nil
}

func createSubnetGroup(svc *rds.RDS, subnetGroupName *string, groupDescription string, subnetIds []string) (*rds.DBSubnetGroup, error) {
	sIds := make([]*string, 0)
	for _, i := range subnetIds {
		sIds = append(sIds, aws.String(i))
	}

	groupInput := &rds.CreateDBSubnetGroupInput{
		DBSubnetGroupName:        subnetGroupName,
		DBSubnetGroupDescription: aws.String(groupDescription),
		SubnetIds:                sIds,
	}

	groupOutput, err := svc.CreateDBSubnetGroup(groupInput)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case rds.ErrCodeDBSubnetGroupAlreadyExistsFault:
				log.Warn(rds.ErrCodeDBSubnetGroupAlreadyExistsFault, aerr.Error())
				return nil, aerr
			case rds.ErrCodeDBSubnetGroupQuotaExceededFault:
				log.Warn(rds.ErrCodeDBSubnetGroupQuotaExceededFault, aerr.Error())
				return nil, aerr
			case rds.ErrCodeDBSubnetQuotaExceededFault:
				log.Warn(rds.ErrCodeDBSubnetQuotaExceededFault, aerr.Error())
				return nil, aerr
			case rds.ErrCodeDBSubnetGroupDoesNotCoverEnoughAZs:
				log.Warn(rds.ErrCodeDBSubnetGroupDoesNotCoverEnoughAZs, aerr.Error())
				return nil, aerr
			case rds.ErrCodeInvalidSubnet:
				log.Warn(rds.ErrCodeInvalidSubnet, aerr.Error())
				return nil, aerr
			default:
				log.Warn(aerr)
				return nil, aerr
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Warn(err)
			return nil, aerr
		}
	}

	return groupOutput.DBSubnetGroup, nil
}

func DeleteDBSubnetGroup(svc *rds.RDS, name string) error {
	input := &rds.DeleteDBSubnetGroupInput{
		DBSubnetGroupName: aws.String(name),
	}

	result, err := svc.DeleteDBSubnetGroup(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case rds.ErrCodeInvalidDBSubnetGroupStateFault:
				log.Warn(rds.ErrCodeInvalidDBSubnetGroupStateFault, aerr.Error())
			case rds.ErrCodeInvalidDBSubnetStateFault:
				log.Warn(rds.ErrCodeInvalidDBSubnetStateFault, aerr.Error())
			case rds.ErrCodeDBSubnetGroupNotFoundFault:
				log.Warn(rds.ErrCodeDBSubnetGroupNotFoundFault, aerr.Error())
			default:
				log.Warn(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Warn(err.Error())
		}
	}

	log.Info(result)
	return nil
}

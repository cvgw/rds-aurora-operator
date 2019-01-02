package cluster

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"
	log "github.com/sirupsen/logrus"
)

var (
	ClusterNotFoundErr error
)

func init() {
	ClusterNotFoundErr = errors.New("cluster not found")
}

func FindDBCluster(svc *rds.RDS, clusterId string) (*rds.DBCluster, error) {
	descClustersInput := &rds.DescribeDBClustersInput{
		DBClusterIdentifier: aws.String(clusterId),
	}

	descClusterOuput, err := svc.DescribeDBClusters(descClustersInput)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case rds.ErrCodeDBClusterNotFoundFault:
				log.Info(rds.ErrCodeDBClusterNotFoundFault, aerr.Error())
				return nil, ClusterNotFoundErr
			default:
				log.Warn(aerr.Error())
				return nil, aerr
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Warn(err.Error())
			return nil, aerr
		}
	}

	return descClusterOuput.DBClusters[0], nil
}

type UpdateDBClusterRequest struct {
	cluster            *rds.DBCluster
	engineVersion      *string
	masterUserPass     *string
	securityGroupIds   []*string
	parameterGroupName *string
}

func (u *UpdateDBClusterRequest) SetCluster(v *rds.DBCluster) *UpdateDBClusterRequest {
	u.cluster = v
	return u
}

func (u *UpdateDBClusterRequest) SetEngineVersion(v string) *UpdateDBClusterRequest {
	u.engineVersion = aws.String(v)
	return u
}

func (u *UpdateDBClusterRequest) SetMasterUserPass(v string) *UpdateDBClusterRequest {
	u.masterUserPass = aws.String(v)
	return u
}

func (u *UpdateDBClusterRequest) SetSecurityGroupIds(v []string) *UpdateDBClusterRequest {
	sIds := make([]*string, 0)
	for _, i := range v {
		sIds = append(sIds, aws.String(i))
	}

	u.securityGroupIds = sIds
	return u
}

func (u *UpdateDBClusterRequest) SetParameterGroupName(v string) *UpdateDBClusterRequest {
	u.parameterGroupName = aws.String(v)
	return u
}

func UpdateDBCluster(svc *rds.RDS, req *UpdateDBClusterRequest) (*rds.DBCluster, error) {
	input := &rds.ModifyDBClusterInput{
		ApplyImmediately:            aws.Bool(true),
		DBClusterIdentifier:         req.cluster.DBClusterIdentifier,
		VpcSecurityGroupIds:         req.securityGroupIds,
		DBClusterParameterGroupName: req.parameterGroupName,
	}

	if *req.cluster.EngineVersion != *req.engineVersion {
		input.EngineVersion = req.engineVersion
	}

	if req.masterUserPass != nil && *req.masterUserPass != "" {
		input.MasterUserPassword = req.masterUserPass
	}

	result, err := svc.ModifyDBCluster(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case rds.ErrCodeDBClusterNotFoundFault:
				log.Warn(rds.ErrCodeDBClusterNotFoundFault, aerr.Error())
				return nil, aerr
			case rds.ErrCodeInvalidDBClusterStateFault:
				log.Warn(rds.ErrCodeInvalidDBClusterStateFault, aerr.Error())
				return nil, aerr
			case rds.ErrCodeStorageQuotaExceededFault:
				log.Warn(rds.ErrCodeStorageQuotaExceededFault, aerr.Error())
				return nil, aerr
			case rds.ErrCodeDBSubnetGroupNotFoundFault:
				log.Warn(rds.ErrCodeDBSubnetGroupNotFoundFault, aerr.Error())
				return nil, aerr
			case rds.ErrCodeInvalidVPCNetworkStateFault:
				log.Warn(rds.ErrCodeInvalidVPCNetworkStateFault, aerr.Error())
				return nil, aerr
			case rds.ErrCodeInvalidDBSubnetGroupStateFault:
				log.Warn(rds.ErrCodeInvalidDBSubnetGroupStateFault, aerr.Error())
				return nil, aerr
			case rds.ErrCodeInvalidSubnet:
				log.Warn(rds.ErrCodeInvalidSubnet, aerr.Error())
				return nil, aerr
			case rds.ErrCodeDBClusterParameterGroupNotFoundFault:
				log.Warn(rds.ErrCodeDBClusterParameterGroupNotFoundFault, aerr.Error())
				return nil, aerr
			case rds.ErrCodeInvalidDBSecurityGroupStateFault:
				log.Warn(rds.ErrCodeInvalidDBSecurityGroupStateFault, aerr.Error())
				return nil, aerr
			case rds.ErrCodeInvalidDBInstanceStateFault:
				log.Warn(rds.ErrCodeInvalidDBInstanceStateFault, aerr.Error())
				return nil, aerr
			case rds.ErrCodeDBClusterAlreadyExistsFault:
				log.Warn(rds.ErrCodeDBClusterAlreadyExistsFault, aerr.Error())
				return nil, aerr
			default:
				log.Warn(aerr.Error())
				return nil, aerr
			}
		} else {
			log.Warn(err)
			return nil, err
		}
	}

	return result.DBCluster, nil
}

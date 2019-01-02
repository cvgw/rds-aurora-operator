package cluster

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	log "github.com/sirupsen/logrus"
)

type NewDBClusterFactoryInput struct {
	ClusterId          string
	Engine             string
	EngineVersion      string
	MasterUsername     string
	MasterUserPass     string
	SecurityGroupIds   []string
	SubnetGroupName    string
	ParameterGroupName string
}

func NewDBClusterFactory(input NewDBClusterFactoryInput) *dbClusterFactory {
	f := &dbClusterFactory{}

	f.clusterIdentifier = aws.String(input.ClusterId)
	f.engine = aws.String(input.Engine)
	f.engineVersion = aws.String(input.EngineVersion)
	f.masterUsername = aws.String(input.MasterUsername)
	f.masterUserPass = aws.String(input.MasterUserPass)

	f.subnetGroupName = aws.String(input.SubnetGroupName)
	f.parameterGroupName = aws.String(input.ParameterGroupName)

	sIds := make([]*string, 0)
	for _, i := range input.SecurityGroupIds {
		sIds = append(sIds, aws.String(i))
	}
	f.securityGroupIds = sIds

	return f
}

type dbClusterFactory struct {
	clusterIdentifier  *string
	subnetGroupName    *string
	securityGroupIds   []*string
	engine             *string
	engineVersion      *string
	masterUsername     *string
	masterUserPass     *string
	parameterGroupName *string
}

func (f *dbClusterFactory) CreateDBCluster(svc *rds.RDS) (*rds.DBCluster, error) {
	clusterInput := &rds.CreateDBClusterInput{
		DBClusterIdentifier:         f.clusterIdentifier,
		Engine:                      f.engine,
		EngineVersion:               f.engineVersion,
		MasterUsername:              f.masterUsername,
		MasterUserPassword:          f.masterUserPass,
		DBSubnetGroupName:           f.subnetGroupName,
		VpcSecurityGroupIds:         f.securityGroupIds,
		DBClusterParameterGroupName: f.parameterGroupName,
	}

	clusterOutput, err := svc.CreateDBCluster(clusterInput)
	if err != nil {
		log.Warn(err)
		return nil, err
	}

	return clusterOutput.DBCluster, nil
}

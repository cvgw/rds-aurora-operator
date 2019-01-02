package cluster

import (
	rdsv1alpha1 "github.com/cvgw/rds-aurora-operator/pkg/apis/rds/v1alpha1"
	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go/service/rds"
	factory "github.com/cvgw/rds-aurora-operator/pkg/lib/factory/cluster"
	clusterProvider "github.com/cvgw/rds-aurora-operator/pkg/lib/provider/cluster"
	"github.com/cvgw/rds-aurora-operator/pkg/lib/service"
	log "github.com/sirupsen/logrus"
)

type CreateClusterRequest struct {
	Spec rdsv1alpha1.ClusterSpec
}

func CreateCluster(svc *rds.RDS, req CreateClusterRequest) (*rds.DBCluster, error) {
	input := factory.NewDBClusterFactoryInput{
		ClusterId:          req.Spec.Id,
		Engine:             req.Spec.Engine,
		EngineVersion:      req.Spec.EngineVersion,
		MasterUsername:     req.Spec.MasterUsername,
		MasterUserPass:     req.Spec.MasterUserPass,
		SecurityGroupIds:   req.Spec.SecurityGroupIds,
		SubnetGroupName:    req.Spec.SubnetGroupName,
		ParameterGroupName: req.Spec.ClusterParameterGroupName,
	}

	clusterFactory := factory.NewDBClusterFactory(input)
	cluster, err := clusterFactory.CreateDBCluster(svc)
	if err != nil {
		return nil, err
	}
	log.Info(cluster)

	return cluster, nil
}

func UpdateDBCluster(svc *rds.RDS, dbCluster *rds.DBCluster, spec rdsv1alpha1.ClusterSpec) error {
	req := &clusterProvider.UpdateDBClusterRequest{}
	req.SetCluster(dbCluster).
		SetEngineVersion(spec.EngineVersion).
		SetSecurityGroupIds(spec.SecurityGroupIds).
		SetParameterGroupName(spec.ClusterParameterGroupName)

	_, err := clusterProvider.UpdateDBCluster(svc, req)
	return err
}

func ValidateCluster(svc *rds.RDS, dbCluster *rds.DBCluster, spec rdsv1alpha1.ClusterSpec) error {
	var err error

	if *dbCluster.DBClusterIdentifier != spec.Id {
		err = service.PopulateValidationErr(err, errors.New("db cluster identifier does not match"))
	}

	if *dbCluster.Engine != spec.Engine {
		err = service.PopulateValidationErr(err, errors.New("db engine does not match"))
	}

	if *dbCluster.EngineVersion != spec.EngineVersion {
		err = service.PopulateValidationErr(err, errors.New("db engine version does not match"))
	}

	if *dbCluster.MasterUsername != spec.MasterUsername {
		err = service.PopulateValidationErr(err, errors.New("db master user name does not match"))
	}

	if *dbCluster.DBSubnetGroup != spec.SubnetGroupName {
		err = service.PopulateValidationErr(err, errors.New("db subnet group name does not match"))
	}

	if spec.ClusterParameterGroupName != "" && *dbCluster.DBClusterParameterGroup != spec.ClusterParameterGroupName {
		err = service.PopulateValidationErr(err, errors.New("db cluster parameter group name does not match"))
	}

	dbSgIds := make([]*string, len(dbCluster.VpcSecurityGroups))
	for i, sg := range dbCluster.VpcSecurityGroups {
		dbSgIds[i] = sg.VpcSecurityGroupId
	}

	if !sgIdsMatch(dbSgIds, spec.SecurityGroupIds) {
		err = service.PopulateValidationErr(err, errors.New("db security groups do not match"))
	}

	return err
}

func sgIdsMatch(dbSgIds []*string, specSgIds []string) bool {
	sg := make([]string, len(dbSgIds))
	for i, sgId := range dbSgIds {
		sg[i] = *sgId
	}
	return service.SliceEqual(sg, specSgIds)
}

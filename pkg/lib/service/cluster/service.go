package cluster

import (
	rdsv1alpha1 "github.com/cvgw/rds-aurora-operator/pkg/apis/rds/v1alpha1"

	"github.com/aws/aws-sdk-go/service/rds"
	factory "github.com/cvgw/rds-aurora-operator/pkg/lib/factory/cluster"
	log "github.com/sirupsen/logrus"
)

type CreateClusterRequest struct {
	Spec rdsv1alpha1.ClusterSpec
}

func CreateCluster(svc *rds.RDS, req CreateClusterRequest) (*rds.DBCluster, error) {
	input := factory.NewDBClusterFactoryInput{
		ClusterId:        req.Spec.Id,
		Engine:           req.Spec.Engine,
		EngineVersion:    req.Spec.EngineVersion,
		MasterUsername:   req.Spec.MasterUsername,
		MasterUserPass:   req.Spec.MasterUserPass,
		SecurityGroupIds: req.Spec.SecurityGroupIds,
		SubnetGroupName:  req.Spec.SubnetGroupName,
	}

	clusterFactory := factory.NewDBClusterFactory(input)
	cluster, err := clusterFactory.CreateDBCluster(svc)
	if err != nil {
		return nil, err
	}
	log.Info(cluster)

	return cluster, nil
}

package cluster

import (
	"context"
	"time"

	rdsv1alpha1 "github.com/cvgw/rds-aurora-operator/pkg/apis/rds/v1alpha1"

	"github.com/aws/aws-sdk-go/service/rds"
	factory "github.com/cvgw/rds-aurora-operator/pkg/lib/factory/cluster"
	provider "github.com/cvgw/rds-aurora-operator/pkg/lib/provider/cluster"
	"github.com/cvgw/rds-aurora-operator/pkg/lib/service"
	log "github.com/sirupsen/logrus"
)

func WaitForClusterReady(ctx context.Context, svc *rds.RDS, cluster *rds.DBCluster) bool {
	var readyCount int

	clusterIdentifier := cluster.DBClusterIdentifier
	for {
		select {
		case <-ctx.Done():
			log.Warn("context expired")
			return false
		default:
			dbCluster, err := provider.FindDBCluster(svc, *clusterIdentifier)
			if err != nil {
				log.Warn(err)
				return false
			}

			if *dbCluster.Status == "available" {
				log.Infof("cluster ready test %d/%d", readyCount+1, service.RequiredReady)
				readyCount++
			} else {
				readyCount = 0
				log.Infof("cluster not ready: status %s", *dbCluster.Status)
			}

			if readyCount == service.RequiredReady {
				log.Info("cluster ready and stable")
				return true
			}

			time.Sleep(service.WaitSleepTime * time.Second)
		}
	}
}

type CreateClusterRequest struct {
	Spec *rdsv1alpha1.ClusterSpec
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

//type CreateClusterRequest struct {
//  ClusterId       string
//  Engine          string
//  EngineVersion   string
//  MasterUsername  string
//  MasterUserPass  string
//  SubnetGroupName string
//  SgIds           []string
//  ReadyTimeout    int
//}

//func CreateCluster(svc *rds.RDS, req CreateClusterRequest) (*rds.DBCluster, error) {
//  clusterFactoryInput := factory.NewDBClusterFactoryInput{
//    ClusterId:        req.ClusterId,
//    Engine:           req.Engine,
//    EngineVersion:    req.EngineVersion,
//    MasterUsername:   req.MasterUsername,
//    MasterUserPass:   req.MasterUserPass,
//    SecurityGroupIds: req.SgIds,
//    SubnetGroupName:  req.SubnetGroupName,
//  }

//  clusterFactory := factory.NewDBClusterFactory(input)
//  cluster, err := clusterFactory.CreateDBCluster(svc)
//  if err != nil {
//    return nil, err
//  }
//  log.Info(cluster)
//  //cluster, err := updateOrCreateCluster(svc, clusterFactoryInput, req.ReadyTimeout)
//  //if err != nil {
//  //  return nil, err
//  //}

//  return cluster, nil
//}

//func updateOrCreateCluster(svc *rds.RDS, input factory.NewDBClusterFactoryInput, rTimeout int) (*rds.DBCluster, error) {
//  clusterFactory := factory.NewDBClusterFactory(input)
//  cluster, err := clusterFactory.UpdateOrCreateDBCluster(svc)
//  if err != nil {
//    return nil, err
//  }
//  log.Info(cluster)

//  ctx, cancel := context.WithTimeout(context.Background(), time.Duration(rTimeout)*time.Minute)
//  defer cancel()
//  ready := factory.WaitForClusterReady(ctx, svc, cluster)

//  if !ready {
//    return nil, errors.New("cluster not ready within timeout")
//  }

//  return cluster, nil
//}

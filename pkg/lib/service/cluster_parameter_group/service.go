package cluster_clusterParameter_group

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"
	rdsv1alpha1 "github.com/cvgw/rds-aurora-operator/pkg/apis/rds/v1alpha1"
	clusterParamGroupProvider "github.com/cvgw/rds-aurora-operator/pkg/lib/provider/cluster_parameter_group"
	"github.com/cvgw/rds-aurora-operator/pkg/lib/service"
	log "github.com/sirupsen/logrus"
)

func CreateClusterParameterGroup(svc *rds.RDS, spec rdsv1alpha1.ClusterParameterGroupSpec) (
	*rds.DBClusterParameterGroup, error,
) {
	req := &clusterParamGroupProvider.CreateRequest{}
	req.SetFamily(spec.Family).
		SetName(spec.Name).
		SetDescription(spec.Description)

	group, err := clusterParamGroupProvider.CreateDBClusterParameterGroup(svc, *req)
	if err != nil {
		return nil, err
	}
	log.Debug(group)

	return group, nil
}

func UpdateClusterParameterGroup(svc *rds.RDS, spec rdsv1alpha1.ClusterParameterGroupSpec) error {
	params := make([]clusterParamGroupProvider.Param, 0)

	for _, p := range spec.Parameters {
		param := clusterParamGroupProvider.Param{
			Apply:     clusterParamGroupProvider.Immediate,
			ValueType: clusterParamGroupProvider.String,
			Name:      p.Name,
			Value:     p.Value,
		}
		params = append(params, param)
	}

	req := &clusterParamGroupProvider.UpdateRequest{}
	req.SetName(spec.Name).
		SetClusterParameters(params)

	return clusterParamGroupProvider.UpdateDBClusterParameterGroup(svc, *req)
}

func ValidateClusterParameterGroup(
	svc *rds.RDS, dbParamGroup *rds.DBClusterParameterGroup, spec rdsv1alpha1.ClusterParameterGroupSpec,
) error {
	var err error

	if *dbParamGroup.DBParameterGroupFamily != spec.Family {
		err = service.PopulateValidationErr(
			err, errors.New("db cluster parameter group family does not match"),
		)
	}

	if *dbParamGroup.Description != spec.Description {
		err = service.PopulateValidationErr(
			err, errors.New("db cluster parameter group description does not match"),
		)
	}

	input := &rds.DescribeDBClusterParametersInput{
		DBClusterParameterGroupName: aws.String(spec.Name),
		MaxRecords:                  aws.Int64(100),
	}

	//result, err := svc.DescribeDBClusterParameters(input)
	paramValidated := make(map[string]map[string]string)
	for _, p := range spec.Parameters {
		m := make(map[string]string)
		m["value"] = p.Value
		m["validated"] = "false"
		paramValidated[p.Name] = m
	}
	validParams := make([]string, 0)
	result, err := svc.DescribeDBClusterParameters(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case rds.ErrCodeDBClusterParameterGroupNotFoundFault:
				log.Warn(rds.ErrCodeDBClusterParameterGroupNotFoundFault, aerr.Error())
			default:
				log.Warn(aerr.Error())
			}
		} else {
			log.Warn(err.Error())
		}
	}
	validParams = append(validParams, validateParams(result.Parameters, paramValidated)...)

	if len(validParams) != len(spec.Parameters) {
		err = service.PopulateValidationErr(
			err, errors.New("db cluster parameter group parameters do not match"),
		)
	}

	return err
}

func validateParams(
	dbParams []*rds.Parameter, paramValidated map[string]map[string]string,
) []string {
	validated := make([]string, 0)
	for _, p := range dbParams {
		for name, data := range paramValidated {
			if *p.ParameterName == name && *p.ParameterValue == data["value"] {
				validated = append(validated, name)
			}
		}
	}

	return validated
}

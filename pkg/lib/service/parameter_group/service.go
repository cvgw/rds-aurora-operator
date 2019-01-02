package parameter_group

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"
	rdsv1alpha1 "github.com/cvgw/rds-aurora-operator/pkg/apis/rds/v1alpha1"
	paramGroupProvider "github.com/cvgw/rds-aurora-operator/pkg/lib/provider/parameter_group"
	"github.com/cvgw/rds-aurora-operator/pkg/lib/service"
	log "github.com/sirupsen/logrus"
)

func CreateParameterGroup(svc *rds.RDS, spec rdsv1alpha1.ParameterGroupSpec) (
	*rds.DBParameterGroup, error,
) {
	req := &paramGroupProvider.CreateRequest{}
	req.SetFamily(spec.Family).
		SetName(spec.Name).
		SetDescription(spec.Description)

	group, err := paramGroupProvider.CreateDBParameterGroup(svc, *req)
	if err != nil {
		return nil, err
	}
	log.Debug(group)

	return group, nil
}

func UpdateParameterGroup(svc *rds.RDS, spec rdsv1alpha1.ParameterGroupSpec) error {
	params := make([]paramGroupProvider.Param, 0)

	for _, p := range spec.Parameters {
		param := paramGroupProvider.Param{
			Apply:     paramGroupProvider.Immediate,
			ValueType: paramGroupProvider.String,
			Name:      p.Name,
			Value:     p.Value,
		}
		params = append(params, param)
	}

	req := &paramGroupProvider.UpdateRequest{}
	req.SetName(spec.Name).
		SetParameters(params)

	return paramGroupProvider.UpdateDBParameterGroup(svc, *req)
}

func ValidateParameterGroup(
	svc *rds.RDS, dbParamGroup *rds.DBParameterGroup, spec rdsv1alpha1.ParameterGroupSpec,
) error {
	var err error

	if *dbParamGroup.DBParameterGroupFamily != spec.Family {
		err = service.PopulateValidationErr(
			err, errors.New("db parameter group family does not match"),
		)
	}

	if *dbParamGroup.Description != spec.Description {
		err = service.PopulateValidationErr(
			err, errors.New("db parameter group description does not match"),
		)
	}

	input := &rds.DescribeDBParametersInput{
		DBParameterGroupName: aws.String(spec.Name),
		MaxRecords:           aws.Int64(100),
	}

	//result, err := svc.DescribeDBParameters(input)
	paramValidated := make(map[string]map[string]string)
	for _, p := range spec.Parameters {
		m := make(map[string]string)
		m["value"] = p.Value
		m["validated"] = "false"
		paramValidated[p.Name] = m
	}
	validParams := make([]string, 0)
	pageNum := 0
	err = svc.DescribeDBParametersPages(
		input,
		func(page *rds.DescribeDBParametersOutput, lastPage bool) bool {
			pageNum++
			validParams = append(validParams, validateParams(page.Parameters, paramValidated)...)

			// Iterate over a maximum of 100 pages
			return pageNum <= 100
		},
	)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case rds.ErrCodeDBParameterGroupNotFoundFault:
				log.Warn(rds.ErrCodeDBParameterGroupNotFoundFault, aerr.Error())
			default:
				log.Warn(aerr.Error())
			}
		} else {
			log.Warn(err.Error())
		}
	}

	if len(validParams) != len(spec.Parameters) {
		err = service.PopulateValidationErr(
			err, errors.New("db parameter group parameters do not match"),
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

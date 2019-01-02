package cluster_parameter_group

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"
	log "github.com/sirupsen/logrus"
)

const (
	Immediate applyMethod = "immediate"
	String    valType     = "string"
)

var (
	NotFoundErr error
)

func init() {
	NotFoundErr = errors.New("db parameter group not found")
}

func FindDBClusterParameterGroup(svc *rds.RDS, paramGroupName string) (*rds.DBClusterParameterGroup, error) {
	input := &rds.DescribeDBClusterParameterGroupsInput{
		DBClusterParameterGroupName: aws.String(paramGroupName),
	}

	result, err := svc.DescribeDBClusterParameterGroups(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case rds.ErrCodeDBParameterGroupNotFoundFault:
				log.Debug(rds.ErrCodeDBParameterGroupNotFoundFault, aerr.Error())
				return nil, NotFoundErr
			default:
				log.Warn(aerr.Error())
				return nil, aerr
			}
		} else {
			log.Warn(err.Error())
			return nil, err
		}
	}
	return result.DBClusterParameterGroups[0], nil
}

type CreateRequest struct {
	family      string
	name        string
	description string
}

func (r *CreateRequest) SetFamily(v string) *CreateRequest {
	r.family = v
	return r
}

func (r *CreateRequest) SetName(v string) *CreateRequest {
	r.name = v
	return r
}

func (r *CreateRequest) SetDescription(v string) *CreateRequest {
	r.description = v
	return r
}

func CreateDBClusterParameterGroup(svc *rds.RDS, req CreateRequest) (
	*rds.DBClusterParameterGroup, error,
) {
	input := &rds.CreateDBClusterParameterGroupInput{
		DBParameterGroupFamily:      aws.String(req.family),
		DBClusterParameterGroupName: aws.String(req.name),
		Description:                 aws.String(req.description),
	}

	result, err := svc.CreateDBClusterParameterGroup(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case rds.ErrCodeDBParameterGroupQuotaExceededFault:
				log.Warn(rds.ErrCodeDBParameterGroupQuotaExceededFault, aerr.Error())
				return nil, aerr
			case rds.ErrCodeDBParameterGroupAlreadyExistsFault:
				log.Warn(rds.ErrCodeDBParameterGroupAlreadyExistsFault, aerr.Error())
				return nil, aerr
			default:
				log.Warn(aerr.Error())
				return nil, aerr
			}
		} else {
			log.Warn(err.Error())
			return nil, err
		}
	}

	return result.DBClusterParameterGroup, nil
}

type UpdateRequest struct {
	name       string
	parameters []*rds.Parameter
}

func (r *UpdateRequest) SetName(v string) *UpdateRequest {
	r.name = v
	return r
}

func (r *UpdateRequest) SetClusterParameters(params []Param) *UpdateRequest {
	awsParams := make([]*rds.Parameter, 0)
	for _, p := range params {
		switch p.ValueType {
		case String:
			awsParams = append(awsParams, &rds.Parameter{
				ApplyMethod:    aws.String(string(p.Apply)),
				ParameterName:  aws.String(p.Name),
				ParameterValue: aws.String(p.Value.(string)),
			})
		}
	}

	r.parameters = awsParams
	return r
}

type applyMethod string
type valType string

type Param struct {
	Apply     applyMethod
	Name      string
	Value     interface{}
	ValueType valType
}

func UpdateDBClusterParameterGroup(svc *rds.RDS, req UpdateRequest) error {
	input := &rds.ModifyDBClusterParameterGroupInput{
		DBClusterParameterGroupName: aws.String(req.name),
		Parameters:                  req.parameters,
		//ClusterParameters: []*rds.ClusterParameter{
		//  {
		//    ApplyMethod:    aws.String("immediate"),
		//    ClusterParameterName:  aws.String("time_zone"),
		//    ClusterParameterValue: aws.String("America/Phoenix"),
		//  },
		//},
	}

	result, err := svc.ModifyDBClusterParameterGroup(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case rds.ErrCodeDBParameterGroupNotFoundFault:
				log.Warn(rds.ErrCodeDBParameterGroupNotFoundFault, aerr.Error())
				return aerr
			case rds.ErrCodeInvalidDBParameterGroupStateFault:
				log.Warn(rds.ErrCodeInvalidDBParameterGroupStateFault, aerr.Error())
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

func DeleteDBClusterParameterGroup(svc *rds.RDS, groupName string) error {
	input := &rds.DeleteDBClusterParameterGroupInput{
		DBClusterParameterGroupName: aws.String(groupName),
	}

	result, err := svc.DeleteDBClusterParameterGroup(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case rds.ErrCodeInvalidDBParameterGroupStateFault:
				log.Warn(rds.ErrCodeInvalidDBParameterGroupStateFault, aerr.Error())
				return err
			case rds.ErrCodeDBClusterParameterGroupNotFoundFault:
				log.Warn(rds.ErrCodeDBClusterParameterGroupNotFoundFault, aerr.Error())
				return NotFoundErr
			default:
				log.Warn(aerr.Error())
				return err
			}
		} else {
			log.Warn(err.Error())
			return err
		}
	}
	log.Debug(result)

	return nil
}

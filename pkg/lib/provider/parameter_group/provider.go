package parameter_group

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

func FindDBParameterGroup(svc *rds.RDS, paramGroupName string) (*rds.DBParameterGroup, error) {
	input := &rds.DescribeDBParameterGroupsInput{
		DBParameterGroupName: aws.String(paramGroupName),
	}

	result, err := svc.DescribeDBParameterGroups(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case rds.ErrCodeDBParameterGroupNotFoundFault:
				log.Info(rds.ErrCodeDBParameterGroupNotFoundFault, aerr.Error())
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
	return result.DBParameterGroups[0], nil
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

func CreateDBParameterGroup(svc *rds.RDS, req CreateRequest) (
	*rds.DBParameterGroup, error,
) {
	input := &rds.CreateDBParameterGroupInput{
		DBParameterGroupFamily: aws.String(req.family),
		DBParameterGroupName:   aws.String(req.name),
		Description:            aws.String(req.description),
	}

	result, err := svc.CreateDBParameterGroup(input)
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

	return result.DBParameterGroup, nil
}

type UpdateRequest struct {
	name       string
	parameters []*rds.Parameter
}

func (r *UpdateRequest) SetName(v string) *UpdateRequest {
	r.name = v
	return r
}

func (r *UpdateRequest) SetParameters(params []Param) *UpdateRequest {
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

func UpdateDBParameterGroup(svc *rds.RDS, req UpdateRequest) error {
	input := &rds.ModifyDBParameterGroupInput{
		DBParameterGroupName: aws.String(req.name),
		Parameters:           req.parameters,
		//Parameters: []*rds.Parameter{
		//  {
		//    ApplyMethod:    aws.String("immediate"),
		//    ParameterName:  aws.String("time_zone"),
		//    ParameterValue: aws.String("America/Phoenix"),
		//  },
		//},
	}

	result, err := svc.ModifyDBParameterGroup(input)
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

func DeleteDBParameterGroup(svc *rds.RDS, groupName string) error {
	input := &rds.DeleteDBParameterGroupInput{
		DBParameterGroupName: aws.String(groupName),
	}

	result, err := svc.DeleteDBParameterGroup(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case rds.ErrCodeInvalidDBParameterGroupStateFault:
				log.Info(rds.ErrCodeInvalidDBParameterGroupStateFault, aerr.Error())
				return err
			case rds.ErrCodeDBParameterGroupNotFoundFault:
				log.Debug(rds.ErrCodeDBParameterGroupNotFoundFault, aerr.Error())
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

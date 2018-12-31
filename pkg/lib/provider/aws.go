package provider

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

const (
	aki     = ""
	sak     = ""
	roleArn = ""
	profile = "dev"
	region  = "us-west-2"
)

//func NewSession(aki, sak, region, roleArn, profile string) *session.Session {
func NewSession() *session.Session {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config:  aws.Config{Region: aws.String(region)},
		Profile: profile,
	}))

	return sess
	//creds := credentials.NewStaticCredentials(aki, sak, "")
	//cfg := aws.NewConfig().WithCredentials(creds).WithRegion(region)
	//sess, err := session.NewSession(cfg)
	//if err != nil {
	//  return err
	//}
	//creds = stscreds.NewCredentials(sess, roleArn)

	//assumeCfg := cfg.Copy()
	//assumeCfg.Credentials = creds

	//assumeSess, err := session.NewSession(assumeCfg)
	//if err != nil {
	//  return err
	//}

	//return assumeSess
}

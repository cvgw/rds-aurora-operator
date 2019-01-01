package environment

import "os"

type AwsSessionEnv struct {
	Aki     string
	Sak     string
	Region  string
	RoleArn string
}

func (env AwsSessionEnv) PopulateEnv() AwsSessionEnv {
	env.Aki = os.Getenv("AWS_ACCESS_KEY_ID")
	env.Sak = os.Getenv("AWS_SECRET_ACCESS_KEY")
	env.Region = os.Getenv("AWS_REGION")
	env.RoleArn = os.Getenv("AWS_ROLE_ARN")
	return env
}

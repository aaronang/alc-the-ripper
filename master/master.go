package master

import (
	// "fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.ReadInConfig()
}

func CreateSlave(svc *ec2.EC2) (*ec2.Instance, error) {
	params := &ec2.RunInstancesInput{
		ImageId:      aws.String(viper.GetString("slave.image")),
		InstanceType: aws.String(viper.GetString("slave.type")),
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(1),
		IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
			Arn: aws.String(viper.GetString("slave.arn")),
		},
	}
	resp, err := svc.RunInstances(params)
	return resp.Instances[0], err
}

func TerminateSlave(svc *ec2.EC2, instance *ec2.Instance) (*ec2.TerminateInstancesOutput, error) {
	params := &ec2.TerminateInstancesInput{
		InstanceIds: []*string{
			instance.InstanceId,
		},
	}
	return svc.TerminateInstances(params)
}

package master

import (
	"bytes"
	"net/http"

	"github.com/aaronang/cong-the-ripper/task"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.ReadInConfig()
}

func CreateSlaves(svc *ec2.EC2, count int64) ([]*ec2.Instance, error) {
	params := &ec2.RunInstancesInput{
		ImageId:      aws.String(viper.GetString("slave.image")),
		InstanceType: aws.String(viper.GetString("slave.type")),
		MinCount:     aws.Int64(count),
		MaxCount:     aws.Int64(count),
		IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
			Arn: aws.String(viper.GetString("slave.arn")),
		},
	}
	resp, err := svc.RunInstances(params)
	return resp.Instances, err
}

func TerminateSlaves(svc *ec2.EC2, instances []*ec2.Instance) (*ec2.TerminateInstancesOutput, error) {
	params := &ec2.TerminateInstancesInput{
		InstanceIds: instanceIds(instances),
	}
	return svc.TerminateInstances(params)
}

func SendTask(t *task.Task, i *ec2.Instance) (resp *http.Response, err error) {
	url := "http://" + *i.PublicIpAddress + ":8080/tasks/create"
	bodyType := "application/json"
	body, _ := t.ToJson()
	return http.Post(url, bodyType, bytes.NewBuffer(body))
}

func instanceIds(instances []*ec2.Instance) []*string {
	instanceIds := make([]*string, len(instances))
	for i, instance := range instances {
		instanceIds[i] = instance.InstanceId
	}
	return instanceIds
}

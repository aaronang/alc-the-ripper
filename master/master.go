package master

import (
	"bytes"
	"net/http"

	"github.com/aaronang/cong-the-ripper/lib"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func CreateSlaves(svc *ec2.EC2, count int64) ([]*ec2.Instance, error) {
	params := &ec2.RunInstancesInput{
		ImageId:      aws.String(lib.SlaveImage),
		InstanceType: aws.String(lib.SlaveType),
		MinCount:     aws.Int64(count),
		MaxCount:     aws.Int64(count),
		IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
			Arn: aws.String(lib.SlaveARN),
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

func SendTask(t *lib.Task, i *ec2.Instance) (*http.Response, error) {
	url := lib.Protocol + *i.PublicIpAddress + lib.Port + lib.CreateTaskPath
	body, err := t.ToJson()
	if err != nil {
		panic(err)
	}
	return http.Post(url, lib.BodyType, bytes.NewBuffer(body))
}

func instanceIds(instances []*ec2.Instance) []*string {
	instanceIds := make([]*string, len(instances))
	for i, instance := range instances {
		instanceIds[i] = instance.InstanceId
	}
	return instanceIds
}

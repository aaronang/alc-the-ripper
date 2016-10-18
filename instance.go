package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

const slaveIPARN = "arn:aws:iam::415077340068:instance-profile/SlaveTheRipper"
const slaveIPName = "SlaveTheRipper"
const masterIPARN = "arn:aws:iam::415077340068:instance-profile/MasterTheRipper"
const masterIPName = "MasterTheRipper"
const region = "eu-west-1"
const imageID = "ami-7abd0209" // CentOS 7 (x86_64) - with Updates HVM for EU (Ireland)
const instanceType = "t2.micro"

func main() {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	if err != nil {
		fmt.Println("Could not create AWS session.", err)
		return
	}

	svc := ec2.New(sess)

	instance, err := createSlave(svc)
	if err != nil {
		fmt.Println("Could not create instance", err)
		return
	}
	fmt.Println(instance)

	terminateSlave(svc, instance)
}

func createSlave(svc *ec2.EC2) (*ec2.Instance, error) {
	runParams := &ec2.RunInstancesInput{
		ImageId:      aws.String(imageID),
		InstanceType: aws.String(instanceType),
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(1),
		IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
			Arn: aws.String(slaveIPARN),
		},
	}
	resp, err := svc.RunInstances(runParams)
	return resp.Instances[0], err
}

func terminateSlave(svc *ec2.EC2, instance *ec2.Instance) (*ec2.TerminateInstancesOutput, error) {
	terminateParams := &ec2.TerminateInstancesInput{
		InstanceIds: []*string{
			instance.InstanceId,
		},
	}
	return svc.TerminateInstances(terminateParams)
}

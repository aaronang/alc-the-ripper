// +build !aws

package master

import (
	"testing"

	"github.com/aaronang/cong-the-ripper/lib"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func TestCreateAndTerminateSlave(t *testing.T) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(lib.AWSRegion)},
	)
	if err != nil {
		t.Error("Could not create AWS session.", err)
	}

	svc := ec2.New(sess)

	instances, err := createSlaves(svc, 3, "8080", "52.49.37.70", "8080")
	if err != nil {
		t.Error("Could not create instance", err)
	}

	if _, err = terminateSlaves(svc, instances); err != nil {
		t.Error("Could not terminate instance", err)
	}
}

func TestCreateSlaveWithUserData(t *testing.T) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(lib.AWSRegion)},
	)
	if err != nil {
		t.Error("Could not create AWS session.", err)
	}

	svc := ec2.New(sess)

	if _, err := createSlaves(svc, 1, "8080", "52.49.37.70", "8080"); err != nil {
		t.Error("Could not create instance", err)
	}
}

func TestSendTask(t *testing.T) {
	ta := &lib.Task{
		JobID:   123,
		Start:   []byte("aaaa"),
		TaskLen: 12,
	}
	if _, err := sendTask(ta, "localhost:8080"); err != nil {
		t.Error("Task did not send correctly", err)
	}
}

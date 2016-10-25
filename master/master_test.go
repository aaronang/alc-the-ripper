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

	instances, err := CreateSlaves(svc, 3)
	if err != nil {
		t.Error("Could not create instance", err)
	}

	if _, err = TerminateSlaves(svc, instances); err != nil {
		t.Error("Could not terminate instance", err)
	}
}

func TestSendTask(t *testing.T) {
	ta := &lib.Task{
		JobID:   123,
		Start:   []byte("aaaa"),
		TaskLen: 12,
	}
	if _, err := SendTask(ta, "localhost"); err != nil {
		t.Error("Task did not send correctly", err)
	}
}

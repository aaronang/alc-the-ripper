// +build !aws

package master

import (
	"encoding/json"
	"fmt"
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

	instances, err := createSlaves(svc, 3)
	if err != nil {
		t.Error("Could not create instance", err)
	}

	if _, err = terminateSlaves(svc, instances); err != nil {
		t.Error("Could not terminate instance", err)
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

func TestStatusMarshal(t *testing.T) {
	lj1 := lib.Job{
		Salt:      []byte("salt"),
		Digest:    []byte("digest"),
		KeyLen:    4,
		Iter:      22,
		Alphabet:  lib.AlphaLower,
		Algorithm: lib.PBKDF2,
	}

	t1 := &lib.Task{
		Job:     lj1,
		JobID:   1,
		ID:      1,
		Start:   []byte("aaaa"),
		TaskLen: 22,
	}

	t2 := &lib.Task{
		Job:     lj1,
		JobID:   1,
		ID:      2,
		Start:   []byte("waa"),
		TaskLen: 22,
	}

	j1 := &job{
		Job:          lj1,
		ID:           1,
		Tasks:        []*lib.Task{t1, t2},
		RunningTasks: 2,
		MaxTasks:     2,
	}

	s1 := slave{
		Tasks:    []*lib.Task{t1, t2},
		MaxSlots: 3,
	}

	status := statusSummary{
		Instances: map[string]slave{
			"52.51.156.198": s1,
		},
		Jobs: []*job{j1},
	}
	fmt.Println(status)
	js, err := json.MarshalIndent(status, "", "\t")
	if err != nil {
		t.Error("Unable to marshal status summary", err)
	}
	fmt.Println(string(js))
}

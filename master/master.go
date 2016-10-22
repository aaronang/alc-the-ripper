package master

import (
	"bytes"
	"net/http"

	"github.com/aaronang/cong-the-ripper/lib"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type slave struct {
	tasks []*lib.Task
	// others
}

type master struct {
	instances map[string]slave
	jobs map[string]*job
	jobsChan chan lib.Job
	heartbeatChan chan lib.Heartbeat
	tasks map[int]*lib.Task
}

func (m *master) Run() {
	http.HandleFunc("/", m.jobsHandler)
	go http.ListenAndServe(lib.Port, nil)
	
	for {
		select {
		case job := <- jobsChan:
			// split the job into tasks
		case beat := <- heartbeatChan:
			// update task statuses
			// check whether a job has completed all its tasks
		}
	}
}

func (m *master)jobsHandler(w http.ResponseWriter, r *http.Request) {
	var j Job
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&j); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	m.jobsChan <- j
}

// CreateSlaves creates a new slave instance.
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

// TerminateSlaves terminates a slave instance.
func TerminateSlaves(svc *ec2.EC2, instances []*ec2.Instance) (*ec2.TerminateInstancesOutput, error) {
	params := &ec2.TerminateInstancesInput{
		InstanceIds: instanceIds(instances),
	}
	return svc.TerminateInstances(params)
}

// SendTask sends a task to a slave instance.
func SendTask(t *lib.Task, i *ec2.Instance) (*http.Response, error) {
	url := lib.Protocol + *i.PublicIpAddress + lib.Port + lib.CreateTaskPath
	body, err := t.ToJSON()
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

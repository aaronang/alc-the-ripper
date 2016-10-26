package master

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/aaronang/cong-the-ripper/lib"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type slave struct {
	tasks    []*lib.Task
	maxTasks int
	// TODO others
}

type scheduler interface {
	schedule(map[string]slave) string
}

type Master struct {
	instances      map[string]slave
	jobs           map[int]*job
	jobsChan       chan lib.Job
	heartbeatChan  chan lib.Heartbeat
	statusChan     chan chan string // dummy
	newTasks       []*lib.Task
	scheduledTasks []*lib.Task
	controllerChan chan string // dummy
	scheduleChan   chan bool   // channel to instruct the main loop to schedule tasks
}

func Init() Master {
	// TODO initialise Master correctly
	return Master{}
}

func (m *Master) Run() {
	http.HandleFunc(lib.JobsCreatePath, m.jobsHandler)
	http.HandleFunc(lib.HeartbeatPath, m.heartbeatHandler)
	http.HandleFunc(lib.StatusPath, m.statusHandler)

	go http.ListenAndServe(lib.Port, nil)
	go func() {
		// TODO Test how this performs when a lot of tasks get submitted.
		time.Sleep(time.Duration(100/(len(m.newTasks)+1)) * time.Millisecond)
		m.scheduleChan <- true
	}()

	for {
		select {
		case <-m.scheduleChan:
			// we shedule the tasks when something is in this channel
			// give the controller new data
			// (controller runs in the background and manages the number of instances)
			// call load balancer function to schedule the tasks
			// move tasks from `newTasks` to `scheduledTasks`
			if slaveIP := m.slaveAvailable(); slaveIP != "" {
				if tidx := m.getTaskToSchedule(); tidx != -1 {
					m.scheduleTask(tidx, slaveIP)
				}
			}
		case job := <-m.jobsChan:
			// split the job into tasks
			// update `jobs` and `newTasks`
			_ = job
		case beat := <-m.heartbeatChan:
			// update task statuses
			// check whether a job has completed all its tasks
			_ = beat
		case c := <-m.statusChan:
			// status handler gives us a channel,
			// we write the status into the channel and the the handler "serves" the result
			_ = c
		}
	}
}

func (m *Master) jobsHandler(w http.ResponseWriter, r *http.Request) {
	var j lib.Job
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&j); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	m.jobsChan <- j
}

func (m *Master) heartbeatHandler(w http.ResponseWriter, r *http.Request) {
	var beat lib.Heartbeat
	// TODO parse json and sends the results directly to the  main loop
	m.heartbeatChan <- beat
}

func (m *Master) statusHandler(w http.ResponseWriter, r *http.Request) {
	resultsChan := make(chan string)
	m.statusChan <- resultsChan
	<-resultsChan
	// TODO read the results and serve status page
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
func SendTask(t *lib.Task, ip string) (*http.Response, error) {
	url := lib.Protocol + ip + lib.Port + lib.TasksCreatePath
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

func (m *Master) getTaskToSchedule() int {
	for idx, t := range m.newTasks {
		if !m.jobs[t.JobID].reachedMaxTasks() {
			return idx
		}
	}
	return -1
}

func (m *Master) slaveAvailable() string {
	minimumTasks := math.MaxInt64
	var slaveIP string
	for ip, i := range m.instances {
		if assignedTasks := len(i.tasks); assignedTasks < minimumTasks && assignedTasks < i.maxTasks {
			minimumTasks = assignedTasks
			slaveIP = ip
		}
	}
	return slaveIP
}

func (m *Master) scheduleTask(tidx int, ip string) {
	if _, err := SendTask(m.newTasks[tidx], ip); err != nil {
		fmt.Println("Sending task to slave did not execute correctly.")
	} else {
		job := m.jobs[m.newTasks[tidx].JobID]
		job.increaseRunningTasks()
		m.scheduledTasks = append(m.scheduledTasks, m.newTasks[tidx])
		m.newTasks = append(m.newTasks[:tidx], m.newTasks[tidx+1:]...)
	}
}

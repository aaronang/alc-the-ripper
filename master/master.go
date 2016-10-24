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
	maxSlots int
}

type scheduler interface {
	schedule(map[string]slave) string
}

type Master struct {
	instances        map[string]slave
	jobs             map[int]*job
	jobsChan         chan lib.Job
	heartbeatChan    chan lib.Heartbeat
	statusChan       chan chan string // dummy
	newTasks         []*lib.Task
	scheduledTasks   []*lib.Task
	controllerTicker *time.Ticker
	scheduleChan     chan bool // channel to instruct the main loop to schedule tasks
	controller       controller
}

type controller struct {
	dt       time.Duration
	kp       float64
	kd       float64
	ki       float64
	prevErr  float64
	integral float64
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
		// TODO test how this performs when a lot of tasks get submitted.
		time.Sleep(time.Duration(100/len(m.newTasks)) * time.Millisecond)
		m.scheduleChan <- true
	}()

	m.controllerTicker = time.NewTicker(m.controller.dt)

	for {
		select {
		case <-m.controllerTicker.C:
			// run one iteration of the controller
			m.runController()
		case <-m.scheduleChan:
			// we shedule the tasks when something is in this channel
			// give the controller new data
			// (controller runs in the background and manages the number of instances)
			// call load balancer function to schedule the tasks
			// move tasks from `newTasks` to `scheduledTasks`
			if m.slotsAvailable() {
				if tIdx := m.getTaskToSchedule(); tIdx != -1 {
					m.scheduleTask(tIdx)
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

func (m *Master) slotsAvailable() bool {
	for _, i := range m.instances {
		if len(i.tasks) < i.maxSlots {
			return true
		}
	}
	return false
}

func (m *Master) scheduleTask(tIdx int) {
	minResources := math.MaxInt64
	var slaveIP string
	for k, v := range m.instances {
		if len(v.tasks) < minResources {
			slaveIP = k
		}
	}
	// NOTE: if SendTask takes too long then it may block the main loop
	if _, err := SendTask(m.newTasks[tIdx], slaveIP); err != nil {
		fmt.Println("Sending task to slave did not execute correctly.")
	} else {
		m.scheduledTasks = append(m.scheduledTasks, m.newTasks[tIdx])
		m.newTasks = append(m.newTasks[:tIdx], m.newTasks[tIdx+1:]...)
	}
}

func (m *Master) countTotalSlots() int {
	cnt := 0
	for _, i := range m.instances {
		cnt += i.maxSlots
	}
	return cnt
}

func (m *Master) maxSlots() int {
	// TODO
	return 20 * 2
}

func (m *Master) countRequiredSlots() int {
	cnt := len(m.scheduledTasks)
	cnt += len(m.newTasks)
	if cnt > m.maxSlots() {
		return m.maxSlots()
	}
	return cnt
}

// runController runs one iteration
func (m *Master) runController() float64 {
	err := float64(m.countRequiredSlots() - m.countTotalSlots())

	dt := m.controller.dt.Seconds()
	m.controller.integral = m.controller.integral + err*dt
	derivative := (err - m.controller.prevErr) / dt
	output := m.controller.kp*err +
		m.controller.ki*m.controller.integral +
		m.controller.kd*derivative
	m.controller.prevErr = err

	return output
}

func (m *Master) killSlaves() {

}

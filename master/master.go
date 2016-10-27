package master

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"time"

	"github.com/aaronang/cong-the-ripper/lib"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type Master struct {
	svc              *ec2.EC2 // safe to be used concurrently
	instances        map[string]slave
	jobs             map[int64]*job
	jobsChan         chan lib.Job
	heartbeatChan    chan lib.Heartbeat
	statusChan       chan chan statusSummary
	newTasks         []*lib.Task
	scheduledTasks   []*lib.Task
	scheduleTicker   *time.Ticker // channel to instruct the main loop to schedule tasks
	controllerTicker *time.Ticker
	controller       controller
	taskSize         int64
	quit             chan bool
}

type slave struct {
	tasks    []*lib.Task
	maxSlots int
	instance *ec2.Instance // can't populate this easily
}

type scheduler interface {
	schedule(map[string]slave) string
}

type controller struct {
	dt       time.Duration
	kp       float64
	kd       float64
	ki       float64
	prevErr  float64
	integral float64
}

type statusSummary struct {
	instances map[string]slave // TODO *ec2.Instance probably won't serialise
	jobs      map[int64]*job
}

// Init creates the master object
func Init() Master {
	// set some defaults
	return Master{
		svc:              nil, // initialised in Run
		instances:        make(map[string]slave),
		jobs:             make(map[int64]*job),
		jobsChan:         make(chan lib.Job),
		heartbeatChan:    make(chan lib.Heartbeat),
		statusChan:       make(chan chan statusSummary),
		newTasks:         make([]*lib.Task, 0),
		scheduledTasks:   make([]*lib.Task, 0),
		scheduleTicker:   nil, // initialised in Run
		controllerTicker: nil, //initialised in Run
		controller: controller{
			dt:       time.Minute * 2,
			kp:       1,
			kd:       0,
			ki:       0,
			prevErr:  0,
			integral: 0,
		},
		taskSize: 6400 * 1000 * 1000,
		quit:     make(chan bool),
	}
}

// Run starts the master
func (m *Master) Run() {
	// initialise the nils
	m.initAWS()
	m.controllerTicker = time.NewTicker(m.controller.dt)
	// TODO test how this performs when a lot of tasks get submitted.
	m.scheduleTicker = time.NewTicker(time.Duration(100/(len(m.newTasks)+1)) * time.Millisecond)

	// setup and run http
	http.HandleFunc(lib.JobsCreatePath, makeJobsHandler(m.jobsChan))
	http.HandleFunc(lib.HeartbeatPath, makeHeartbeatHandler(m.heartbeatChan))
	http.HandleFunc(lib.StatusPath, makeStatusHandler(m.statusChan))
	go func() {
		e := http.ListenAndServe(lib.Port, nil)
		if e != nil {
			fmt.Println(e)
		}
	}()

	// send message to m.quit on interrupt
	go func() {
		sigchan := make(chan os.Signal, 10)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan
		m.quit <- true
	}()

	// main loop
	for {
		select {
		case <-m.controllerTicker.C:
			// run one iteration of the controller
			m.runController()
		case <-m.scheduleTicker.C:
			// we shedule the tasks when something is in this channel
			// (controller runs in the background and manages the number of instances)
			// call load balancer function to schedule the tasks
			// move tasks from `newTasks` to `scheduledTasks`
			if slaveIP := m.slaveAvailable(); slaveIP != "" {
				if tidx := m.getTaskToSchedule(); tidx != -1 {
					m.scheduleTask(tidx, slaveIP)
				}
			}
		case j := <-m.jobsChan:
			// split the job into tasks
			newJob := job{
				Job:          j,
				id:           rand.Int63(),
				runningTasks: 0,
				maxTasks:     0,
			}
			newJob.splitJob(m.taskSize)

			// update `jobs` and `newTasks`
			m.jobs[newJob.id] = &newJob
			for i := range newJob.tasks {
				m.newTasks = append(m.newTasks, newJob.tasks[i])
			}
		case beat := <-m.heartbeatChan:
			// update task statuses
			// check whether a job has completed all its tasks
			m.updateOnHeartbeat(beat)
		case s := <-m.statusChan:
			// status handler gives us a channel,
			// we write the status into the channel and the the handler serves the result
			_ = s
		case <-m.quit:
			// release all slaves
			fmt.Println("Master stopping...")
			_, err := terminateSlaves(m.svc, slavesMapToInstances(m.instances))
			if err != nil {
				fmt.Println("Failed to terminate slaves on interrupt")
				fmt.Println(err)
			}
			os.Exit(0)
		}
	}
}

// createSlaves creates a new slave instance.
func createSlaves(svc *ec2.EC2, count int64) ([]*ec2.Instance, error) {
	params := &ec2.RunInstancesInput{
		ImageId:      aws.String(lib.SlaveImage),
		InstanceType: aws.String(lib.SlaveType),
		MinCount:     aws.Int64(count),
		MaxCount:     aws.Int64(count),
		IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
			Arn: aws.String(lib.SlaveARN),
		},
		SecurityGroupIds: []*string{aws.String("sg-646fbb02")},
	}
	resp, err := svc.RunInstances(params)
	return resp.Instances, err
}

// terminateSlaves terminates a slave instance.
func terminateSlaves(svc *ec2.EC2, instances []*ec2.Instance) (*ec2.TerminateInstancesOutput, error) {
	params := &ec2.TerminateInstancesInput{
		InstanceIds: instanceIds(instances),
	}
	return svc.TerminateInstances(params)
}

// sendTask sends a task to a slave instance.
func sendTask(t *lib.Task, ip string) (*http.Response, error) {
	url := lib.Protocol + ip + lib.Port + lib.TasksCreatePath
	body, err := t.ToJSON()
	if err != nil {
		panic(err)
	}
	return http.Post(url, lib.BodyType, bytes.NewBuffer(body))
}

func newEC2() *ec2.EC2 {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(lib.AWSRegion)},
	)
	if err != nil {
		panic(err)
	}
	return ec2.New(sess)
}

func (m *Master) initAWS() {
	m.svc = newEC2()

	// create one slave on startup
	s, err := createSlaves(m.svc, 1)
	if err != nil {
		fmt.Println("Failed to create slave")
		panic(err)
	}

	ip := getPublicIP(m.svc, s[0])
	// probably not necessary becase
	if ip != nil {
		m.instances[*ip] = slave{
			maxSlots: lib.MaxSlotsPerInstance,
			instance: s[0],
		}
	}
}

func getPublicIP(svc *ec2.EC2, instance *ec2.Instance) *string {
	params := ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("instance-state-name"),
				Values: []*string{
					aws.String("pending"),
					aws.String("running"),
				},
			},
		},
		InstanceIds: []*string{
			instance.InstanceId,
		},
	}

	var i int
	for {
		res, err := svc.DescribeInstances(&params)
		fmt.Println("reservations", len(res.Reservations))

		// ignore the error because we may try again
		if err == nil &&
			len(res.Reservations) == 1 &&
			len(res.Reservations[0].Instances) == 1 {

			fmt.Println(res.Reservations[0].Instances[0].State)
			if res.Reservations[0].Instances[0].PublicIpAddress != nil {
				fmt.Println("Found public IP")
				return res.Reservations[0].Instances[0].PublicIpAddress
			}

		}
		time.Sleep(10 * time.Second)
		i++
		if i > 12 {
			fmt.Println("Unable to fine public IP")
			return nil
		}
	}
}

func makeJobsHandler(c chan lib.Job) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var j lib.Job
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&j); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		c <- j
	}
}

func makeHeartbeatHandler(c chan lib.Heartbeat) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var beat lib.Heartbeat
		// TODO parse json and sends the results directly to the main loop
		c <- beat
	}
}

func makeStatusHandler(c chan chan statusSummary) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resultsChan := make(chan statusSummary)
		c <- resultsChan
		<-resultsChan
		// TODO read the results and serve status page
	}
}

func (m *Master) updateOnHeartbeat(beat lib.Heartbeat) {
	// check whether slave already exist, if not create one
	// create a goroutine for every slave that checks for missed hearbeats
	// if a miss is detected then report to master to be handled
	// update task statuses on every heartbeat
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
		if assignedTasks := len(i.tasks); assignedTasks < minimumTasks && assignedTasks < i.maxSlots {
			minimumTasks = assignedTasks
			slaveIP = ip
		}
	}
	return slaveIP
}

func (m *Master) countTotalSlots() int {
	cnt := 0
	for _, i := range m.instances {
		cnt += i.maxSlots
	}
	return cnt
}

func (m *Master) maxSlots() int {
	// TODO do we set the manually or it's a property of AWS?
	return 20 * lib.MaxSlotsPerInstance
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
func (m *Master) runController() {
	err := float64(m.countRequiredSlots() - m.countTotalSlots())

	dt := m.controller.dt.Seconds()
	m.controller.integral = m.controller.integral + err*dt
	derivative := (err - m.controller.prevErr) / dt
	output := m.controller.kp*err +
		m.controller.ki*m.controller.integral +
		m.controller.kd*derivative
	m.controller.prevErr = err

	fmt.Printf("err: %v, output: %v\n", err, output)
	m.adjustInstanceCount(int(output))
}

func (m *Master) adjustInstanceCount(n int) {
	if n > 0 {
		go func() {
			_, err := createSlaves(m.svc, int64(n))
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Printf("%v instances created successfully\n", n)
				// no need to report back to the master loop
				// because it should start receiving heartbeat messages
			}
		}()
	} else {
		// negate n to represent the (positive) number of instances to kill
		// scale by the number of
		n := -n
		n = n / lib.MaxSlotsPerInstance
		if n < 0 {
			panic("n cannot be negative")
		} else if n == 0 {
			fmt.Println("n is 0 in killSlaves")
			return
		}

		// kills n least loaded slaves, the killed slaves may have unfinished tasks
		// but the master should detect missing heartbeats and restart the tasks
		slaves := make([]slave, len(m.instances))
		var i int
		for _, v := range m.instances {
			slaves[i] = v
			i++
		}

		sort.Sort(byTaskCount(slaves))
		go func() {
			_, err := terminateSlaves(m.svc, slavesToInstances(slaves[:n]))
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Printf("%v instances terminated successfully", n)
				// again, no need to report success/failure
				// because heartbeat messages will stop
			}
		}()
	}
}

type byTaskCount []slave

func (a byTaskCount) Len() int {
	return len(a)
}

func (a byTaskCount) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a byTaskCount) Less(i, j int) bool {
	return len(a[i].tasks) < len(a[j].tasks)
}

func slavesToInstances(slaves []slave) []*ec2.Instance {
	res := make([]*ec2.Instance, len(slaves))
	for i := range slaves {
		res[i] = slaves[i].instance
	}
	return res
}

func slavesMapToInstances(slaves map[string]slave) []*ec2.Instance {
	res := make([]*ec2.Instance, len(slaves))
	i := 0
	for _, v := range slaves {
		res[i] = v.instance
	}
	return res
}

func (m *Master) scheduleTask(tidx int, ip string) {
	// NOTE: if sendTask takes too long then it may block the main loop
	if _, err := sendTask(m.newTasks[tidx], ip); err != nil {
		fmt.Println("Sending task to slave did not execute correctly.")
	} else {
		job := m.jobs[m.newTasks[tidx].JobID]
		job.increaseRunningTasks()
		m.scheduledTasks = append(m.scheduledTasks, m.newTasks[tidx])
		m.newTasks = append(m.newTasks[:tidx], m.newTasks[tidx+1:]...)
	}
}

package master

import (
	"encoding/json"
	"log"
	"math"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/aaronang/cong-the-ripper/lib"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type Master struct {
	port              string
	svc               *ec2.EC2 // safe to be used concurrently
	instances         map[string]slave
	jobs              map[int]*job
	jobsChan          chan lib.Job
	heartbeatChan     chan heartbeat
	heartbeatMissChan chan string // represents a key to instances
	statusChan        chan chan statusSummary
	newTasks          []*lib.Task
	scheduledTasks    []*lib.Task
	scheduleTicker    *time.Ticker // channel to instruct the main loop to schedule tasks
	controllerTicker  *time.Ticker
	controller        controller
	taskSize          int
	quit              chan bool
}

type slave struct {
	tasks    []*lib.Task
	maxSlots int
	instance *ec2.Instance // can't populate this easily
	c        chan<- bool
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

type heartbeat struct {
	lib.Heartbeat
	addr string
}

// Init creates the master object
func Init(port string) Master {
	// set some defaults
	return Master{
		port:              port,
		svc:               nil, // initialised in Run
		instances:         make(map[string]slave),
		jobs:              make(map[int]*job),
		jobsChan:          make(chan lib.Job),
		heartbeatChan:     make(chan heartbeat),
		heartbeatMissChan: make(chan string),
		statusChan:        make(chan chan statusSummary),
		newTasks:          make([]*lib.Task, 0),
		scheduledTasks:    make([]*lib.Task, 0),
		scheduleTicker:    nil, // initialised in Run
		controllerTicker:  nil, //initialised in Run
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
	go func() {
		http.HandleFunc(lib.JobsCreatePath, makeJobsHandler(m.jobsChan))
		http.HandleFunc(lib.HeartbeatPath, makeHeartbeatHandler(m.heartbeatChan))
		http.HandleFunc(lib.StatusPath, makeStatusHandler(m.statusChan))
		log.Println("Running master on port", m.port)
		e := http.ListenAndServe(":"+m.port, nil)
		if e != nil {
			log.Fatalln(e)
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
				id:           rand.Int(),
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
		case addr := <-m.heartbeatMissChan:
			// do something when heartbeat is missed
			_ = addr
		case s := <-m.statusChan:
			// status handler gives us a channel,
			// we write the status into the channel and the the handler serves the result
			_ = s
		case <-m.quit:
			// release all slaves
			log.Println("Master stopping...")
			_, err := terminateSlaves(m.svc, slavesMapToInstances(m.instances))
			if err != nil {
				log.Fatalln("Failed to terminate slaves on interrupt", err)
			}
			os.Exit(0)
		}
	}
}

func (m *Master) initAWS() {
	m.svc = newEC2()

	// create one slave on startup
	s, err := createSlaves(m.svc, 1)
	if err != nil {
		log.Fatalln("Failed to create slave", err)
	}

	// NOTE: not necessary when heartbeat message are working
	// master should update its fields according to the heartbeats
	ip := getPublicIP(m.svc, s[0])
	if ip != nil {
		m.instances[*ip] = slave{
			maxSlots: lib.MaxSlotsPerInstance,
			instance: s[0],
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

func makeHeartbeatHandler(c chan heartbeat) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var beat lib.Heartbeat
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&beat); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		c <- heartbeat{
			Heartbeat: beat,
			addr:      r.RemoteAddr,
		}
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

func (m *Master) updateTask(status lib.TaskStatus, ip string) {
	for i := range m.jobs[status.JobId].tasks {
		task := m.jobs[status.JobId].tasks[i]
		if task.ID == status.Id {
			if status.Status == lib.Running {
				task.Progress = status.Progress
			} else {
				if status.Status == lib.PasswordFound {
					// TODO terminate the other tasks in the same job if a password is found
					log.Printf("Password found: %v (task: %v, job, %v)\n",
						status.Password, status.Id, status.JobId)
				}
				m.removeTask(ip, status.JobId, status.Id)

				// remove the job if it's finished
				if len(m.jobs[status.JobId].tasks) == 0 {
					delete(m.jobs, status.JobId)
					log.Printf("Job %v completed", status.JobId)
				}

			}
		}
	}
}

func (m *Master) removeTask(ip string, jobID, taskID int) {
	jobIdx := -1
	for i := range m.jobs[jobID].tasks {
		if m.jobs[jobID].tasks[i].ID == taskID {
			jobIdx = i
			break
		}
	}

	scheduledIdx := -1
	for i := range m.scheduledTasks {
		if m.scheduledTasks[i].JobID == jobID && m.scheduledTasks[i].ID == taskID {
			scheduledIdx = i
			break
		}
	}

	instanceIdx := -1
	for i := range m.instances[ip].tasks {
		t := m.instances[ip].tasks[i]
		if t.ID == taskID && t.JobID == jobID {
			instanceIdx = i
			break
		}
	}

	if jobIdx != -1 && scheduledIdx != -1 && instanceIdx != -1 {
		m.jobs[jobID].tasks = append(m.jobs[jobID].tasks[:jobIdx], m.jobs[jobID].tasks[jobIdx+1:]...)
		m.scheduledTasks = append(m.scheduledTasks[:scheduledIdx], m.scheduledTasks[scheduledIdx+1:]...)

		tmp := m.instances[ip]
		tmp.tasks = append(tmp.tasks[:instanceIdx], tmp.tasks[instanceIdx+1:]...)
		m.instances[ip] = tmp
	} else {
		log.Fatalf("Inconsistent behaviour in removeTask - jobIdx: %v, scheduledIdx: %v, instanceIdx: %v",
			jobIdx, scheduledIdx, instanceIdx)
	}
}

func (m *Master) updateOnHeartbeat(beat heartbeat) {
	if _, ok := m.instances[beat.addr]; ok {
		// existing slave, update
		// update task statuses on every heartbeat
		for _, s := range beat.TaskStatus {
			m.updateTask(s, beat.addr)
		}
	} else {
		// NOTE: this function should not block the main loop
		// consider changing it if instancesFromIPs take too long
		instance := instancesFromIPs(m.svc, []string{beat.addr})[0]
		m.instances[beat.addr] = slave{
			tasks:    make([]*lib.Task, 0),
			maxSlots: lib.MaxSlotsPerInstance,
			instance: instance,
			c:        heartbeatChecker(beat.addr, m.heartbeatMissChan),
		}
		log.Printf("New instance %v created.", beat.addr)
	}
}

func heartbeatChecker(addr string, missedChan chan<- string) chan<- bool {
	beatChan := make(chan bool)
	go func() {
		for {
			timeout := time.After(5 * time.Second)
			select {
			case <-timeout:
				log.Printf("Missed heartbeat for %v", addr)
				missedChan <- addr
				return
			case <-beatChan:
				// ok, do nothing
			}
		}
	}()
	return beatChan
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

func (m *Master) scheduleTask(tidx int, ip string) {
	// NOTE: if sendTask takes too long then it may block the main loop
	if _, err := sendTask(m.newTasks[tidx], net.JoinHostPort(ip, m.port)); err != nil {
		log.Println("Sending task to slave did not execute correctly.", err)
	} else {
		job := m.jobs[m.newTasks[tidx].JobID]
		job.increaseRunningTasks()
		m.scheduledTasks = append(m.scheduledTasks, m.newTasks[tidx])
		m.newTasks = append(m.newTasks[:tidx], m.newTasks[tidx+1:]...)
	}
}

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
	port             string
	svc              *ec2.EC2 // safe to be used concurrently
	instances        map[string]slave
	jobs             map[int]*job
	jobsChan         chan lib.Job
	heartbeatChan    chan lib.Heartbeat
	statusChan       chan chan statusSummary
	newTasks         []*lib.Task
	scheduledTasks   []*lib.Task
	scheduleTicker   *time.Ticker // channel to instruct the main loop to schedule tasks
	controllerTicker *time.Ticker
	controller       controller
	taskSize         int
	quit             chan bool
}

type slave struct {
	Tasks    []*lib.Task   `json:"tasks"`
	MaxSlots int           `json:"maxSlots"`
	Instance *ec2.Instance `json:"-"` // can't populate this easily
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
	Instances map[string]slave `json:"instances"` // TODO *ec2.Instance probably won't serialise
	Jobs      []*job           `json:"jobs"`
}

// Init creates the master object
func Init(port string) Master {
	// set some defaults
	return Master{
		port:             port,
		svc:              nil, // initialised in Run
		instances:        make(map[string]slave),
		jobs:             make(map[int]*job),
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
				ID:           rand.Int(),
				RunningTasks: 0,
				MaxTasks:     0,
			}
			newJob.splitJob(m.taskSize)

			// update `jobs` and `newTasks`
			m.jobs[newJob.ID] = &newJob
			for i := range newJob.Tasks {
				m.newTasks = append(m.newTasks, newJob.Tasks[i])
			}
		case beat := <-m.heartbeatChan:
			// update task statuses
			// check whether a job has completed all its tasks
			m.updateOnHeartbeat(beat)
		case s := <-m.statusChan:
			// status handler gives us a channel,
			// we write the status into the channel and the handler serves the result
			var jobs []*job
			for i := range m.jobs {
				jobs = append(jobs, m.jobs[i])
			}
			s <- statusSummary{
				Instances: m.instances,
				Jobs:      jobs,
			}
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
			MaxSlots: lib.MaxSlotsPerInstance,
			Instance: s[0],
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
		res, err := json.MarshalIndent(<-resultsChan, "", "\t")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(res)
	}
}

func (m *Master) updateOnHeartbeat(beat lib.Heartbeat) {
	// check whether slave already exist, if not create one
	// create a goroutine for every slave that checks for missed hearbeats
	// if a miss is detected then report to master to be handled
	// update task statuses on every heartbeat
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
		if assignedTasks := len(i.Tasks); assignedTasks < minimumTasks && assignedTasks < i.MaxSlots {
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

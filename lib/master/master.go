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
	"strconv"
	"time"

	"github.com/aaronang/cong-the-ripper/lib"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type Master struct {
	ip                string
	port              string
	svc               *ec2.EC2 // safe to be used concurrently
	instances         map[string]slave
	jobs              map[int]*job
	completedJobs     map[int]*job
	jobsChan          chan lib.Job
	heartbeatChan     chan heartbeat
	heartbeatMissChan chan string // represents a key to instances
	statusChan        chan chan StatusJSON
	newTasks          []*lib.Task
	scheduledTasks    []*lib.Task
	scheduleTicker    *time.Ticker // channel to instruct the main loop to schedule tasks
	controllerTicker  *time.Ticker
	controller        controller
	taskSize          int
	quit              chan bool
}

type slave struct {
	tasks         []*lib.Task
	maxSlots      int
	heartbeatChan chan<- bool
	killChan      chan<- int
}

type controller struct {
	dt       time.Duration
	kp       float64
	kd       float64
	ki       float64
	prevErr  float64
	integral float64
}

type heartbeat struct {
	lib.Heartbeat
	ip string
}

// Init creates the master object
func Init(port, ip string, kp, ki, kd float64) Master {
	// set some defaults
	return Master{
		ip:                ip,
		port:              port,
		svc:               nil, // initialised in Run
		instances:         make(map[string]slave),
		jobs:              make(map[int]*job),
		completedJobs:     make(map[int]*job),
		jobsChan:          make(chan lib.Job),
		heartbeatChan:     make(chan heartbeat),
		heartbeatMissChan: make(chan string),
		statusChan:        make(chan chan StatusJSON),
		newTasks:          make([]*lib.Task, 0),
		scheduledTasks:    make([]*lib.Task, 0),
		scheduleTicker:    nil, // initialised in Run
		controllerTicker:  nil, //initialised in Run
		controller: controller{
			dt:       time.Minute*2 + time.Second*30,
			kp:       kp,
			kd:       kd,
			ki:       ki,
			prevErr:  0,
			integral: 0,
		},
		// the amount of raw hashes to compute to achieve ~10 minute duration
		// but tasks get killed when the result is founds, the average duration should be halved
		taskSize: 284000000,
		quit:     make(chan bool),
	}
}

// Run starts the master
func (m *Master) Run() {
	// initialise the nils
	m.svc = newEC2()
	m.controllerTicker = time.NewTicker(m.controller.dt)
	// TODO test how this performs when a lot of tasks get submitted.
	m.scheduleTicker = time.NewTicker(time.Duration(100/(len(m.newTasks)+1)) * time.Millisecond)

	// setup and run http
	go func() {
		http.HandleFunc(lib.JobsCreatePath, makeJobsHandler(m.jobsChan))
		http.HandleFunc(lib.HeartbeatPath, makeHeartbeatHandler(m.heartbeatChan))
		http.HandleFunc(lib.StatusPath, makeStatusHandler(m.statusChan))
		log.Println("[Run] Running master on port", m.port)
		e := http.ListenAndServe(":"+m.port, nil)
		if e != nil {
			log.Panicln("[Run]", e)
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
				maxTasks:     4, // TODO decide this value
				startTime:    time.Now(),
				// we keep finishTime to the zero value
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
			m.instances[beat.ip].heartbeatChan <- true

		case ip := <-m.heartbeatMissChan:
			// moved the scheduled tasks back to new tasks to be re-scheduled
			for i := range m.instances[ip].tasks {
				task := m.instances[ip].tasks[i]
				m.scheduledTasks = removeTaskFrom(m.scheduledTasks, task.JobID, task.ID)
				m.newTasks = append([]*lib.Task{task}, m.newTasks...)
				m.jobs[task.JobID].decreaseRunningTasks()
			}
			delete(m.instances, ip)
		case s := <-m.statusChan:
			// status handler gives us a channel,
			// we write the status into the channel and the handler serves the result
			s <- createStatusJSON(m)
		case <-m.quit:
			// release all slaves
			log.Println("[Run] Master stopping...")
			instances := instancesFromIPs(m.svc, mapToKeys(m.instances))
			if instances != nil && len(instances) > 0 {
				_, err := terminateSlaves(m.svc, instances)
				if err != nil {
					log.Panicln("[Run] Failed to terminate slaves on interrupt", err)
				}

			}
			os.Exit(0)
		}
	}
}

func makeJobsHandler(c chan lib.Job) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("[JobsHandler] received request")
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
		log.Println("[HeartbeatHandler] received request")
		var beat lib.Heartbeat
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&beat); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		c <- heartbeat{
			Heartbeat: beat,
			ip:        ip,
		}
	}
}

func makeStatusHandler(c chan chan StatusJSON) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("[StatusHandler] received request")
		statusChan := make(chan StatusJSON)
		c <- statusChan
		res, err := json.MarshalIndent(<-statusChan, "", "\t")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(res)
		if err != nil {
			log.Println("[StatusHandler] Failed to write status", err)
		}
	}
}

func (m *Master) updateTask(status lib.TaskStatus, ip string) {
	if _, ok := m.jobs[status.JobId]; !ok {
		log.Println("[updateTask] there is a job we don't know about", status.JobId)
		// NOTE behaviour is undefine if instances are running tasks/jobs that we don't know about
		// TODO kill that job/task
		return
	}

	for i := range m.jobs[status.JobId].tasks {
		task := m.jobs[status.JobId].tasks[i]
		if task.ID == status.Id {
			if status.Status == lib.Running {
				// updating this pointer should be enough,
				// it should already exist in the instances and scheduledTasks fields
				task.Progress = status.Progress
			} else {
				if status.Status == lib.PasswordFound {
					// TODO terminate the other tasks in the same job if a password is found
					log.Printf("[updateTask] Password found!!!!: %v (task: %v, job, %v)\n",
						status.Password, status.Id, status.JobId)
					tmpJob := m.jobs[status.JobId]
					tmpJob.password = status.Password
					m.jobs[status.JobId] = tmpJob

					m.clearNewTaskOfJob(status.JobId)
					for _, inst := range m.instances {
						inst.killChan <- status.JobId
					}
				} else {
					log.Printf("[updateTask] Password not found: %v (task: %v, job, %v)\n",
						status.Password, status.Id, status.JobId)
				}

				m.removeTask(ip, status.JobId, status.Id)

				// move to completedJobs for reporting purposes
				if len(m.jobs[status.JobId].tasks) == 0 {
					j := m.jobs[status.JobId]
					j.finishTime = time.Now()
					m.completedJobs[status.JobId] = j

					delete(m.jobs, status.JobId)

					log.Printf("[updateTask] Job %v completed at %v", status.JobId, j.finishTime)
				}
			}
			break
		}
	}
}

func (m *Master) clearNewTaskOfJob(jobID int) {
	for {
		for i := range m.newTasks {
			if m.newTasks[i].JobID == jobID {
				m.newTasks = append(m.newTasks[:i], m.newTasks[i+1:]...)
				break
			}
		}
		break
	}
}

func (m *Master) killTasksOnSlave(jobID int) {
	for ip, inst := range m.instances {
		for i := range inst.tasks {
			if inst.tasks[i].JobID == jobID {
				// TODO why are we using master's port?
				addr := net.JoinHostPort(ip, m.port)
				jobIDStr := strconv.Itoa(jobID)
				_, err := http.Get(lib.Protocol + addr + lib.JobsKillPath + "?jobid=" + jobIDStr)
				if err != nil {
					log.Panicln("[killTasksOnSlave] failed send kill job request", err)
				}

				// only send one request per instance
				break
			}
		}
	}
}

func sendKillRequest(addr string, jobID int) {
	jobIDStr := strconv.Itoa(jobID)
	_, err := http.Get(lib.Protocol + addr + lib.JobsKillPath + "?jobid=" + jobIDStr)
	if err != nil {
		log.Panicln("[sendKillRequest] failed send kill job request", err)
	}
}

// removeTaskFrom returns a slice of task pointers as the result of removal
func removeTaskFrom(tasks []*lib.Task, jobID, taskID int) []*lib.Task {
	for i := range tasks {
		if tasks[i].ID == taskID && tasks[i].JobID == jobID {
			return append(tasks[:i], tasks[i+1:]...)
		}
	}
	return nil
}

func (m *Master) removeTask(ip string, jobID, taskID int) {
	jobsRes := removeTaskFrom(m.jobs[jobID].tasks, jobID, taskID)
	scheduledRes := removeTaskFrom(m.scheduledTasks, jobID, taskID)
	instancesRes := removeTaskFrom(m.instances[ip].tasks, jobID, taskID)

	if jobsRes != nil && scheduledRes != nil && instancesRes != nil {
		log.Printf("[removeTask] job: %v, task: %v", jobID, taskID)
		m.jobs[jobID].tasks = jobsRes
		m.jobs[jobID].decreaseRunningTasks()

		m.scheduledTasks = scheduledRes

		tmp := m.instances[ip]
		tmp.tasks = instancesRes
		m.instances[ip] = tmp
	} else {
		log.Printf("[removeTask] Failed to removed task - job: %v, task: %v\n", jobID, taskID)
	}
}

func (m *Master) updateOnHeartbeat(beat heartbeat) {
	if _, ok := m.instances[beat.ip]; ok { // for instances that already exist
		log.Println("[updateOnHeartbeat] updating existing instance", beat.ip)
		for _, s := range beat.TaskStatus {
			m.updateTask(s, beat.ip)
		}
	} else { // for new instances
		hbc, kc := heartbeatChecker(beat.ip, m.port, m.heartbeatMissChan)
		m.instances[beat.ip] = slave{
			tasks:         make([]*lib.Task, 0),
			maxSlots:      lib.MaxSlotsPerInstance,
			heartbeatChan: hbc,
			killChan:      kc,
		}
		log.Printf("[updateOnHeartbeat] New instance %v created.", beat.ip)
	}
}

func heartbeatChecker(ip, port string, missedChan chan<- string) (chan<- bool, chan<- int) {
	beatChan := make(chan bool)
	killChan := make(chan int, 100)
	go func() {
		for {
			timeout := time.After(2 * lib.HeartbeatInterval)
			select {
			case <-timeout:
				log.Printf("[heartbeatChecker] Missed heartbeat for %v", ip)
				missedChan <- ip
				return
			case <-beatChan:
			inner:
				// heartbeate ok, check if there are jobs that we need to kill
				for {
					select {
					case jobID := <-killChan:
						sendKillRequest(net.JoinHostPort(ip, port), jobID)
					default:
						break inner
					}
				}
			}
		}
	}()
	return beatChan, killChan
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
	log.Println("[scheduleTask] Scheduling task to", ip)
	// NOTE: if sendTask takes too long then it may block the main loop, and why are we using master's port?
	if _, err := sendTask(m.newTasks[tidx], net.JoinHostPort(ip, m.port)); err != nil {
		log.Println("[scheduleTask] Sending task to slave did not execute correctly.", err)
	} else {
		log.Printf("[scheduleTask] scheduled new task %v to %v\n", m.newTasks[tidx].ID, ip)
		job := m.jobs[m.newTasks[tidx].JobID]
		job.increaseRunningTasks()

		inst := m.instances[ip]
		inst.tasks = append(inst.tasks, m.newTasks[tidx])
		m.instances[ip] = inst

		m.scheduledTasks = append(m.scheduledTasks, m.newTasks[tidx])
		m.newTasks = append(m.newTasks[:tidx], m.newTasks[tidx+1:]...)
	}
}

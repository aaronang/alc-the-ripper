package slave

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/aaronang/cong-the-ripper/lib"
)

var slaveInstance Slave

type task struct {
	lib.Task
	Status       lib.Status
	Password     string
	progressChan chan chan []byte
	killChan     chan bool
}

type Slave struct {
	port            string
	masterIp        string
	masterPort      string
	successChan     chan CrackerSuccess
	failChan        chan CrackerFail
	addTaskChan     chan lib.Task
	killChan        chan int
	heartbeatTicker *time.Ticker
	tasks           []*task
}

func Init(port, masterIp, masterPort string) *Slave {
	slaveInstance = Slave{
		port:            port,
		masterIp:        masterIp,
		masterPort:      masterPort,
		successChan:     make(chan CrackerSuccess),
		failChan:        make(chan CrackerFail),
		addTaskChan:     make(chan lib.Task),
		killChan:        make(chan int),
		heartbeatTicker: nil,
	}
	return &slaveInstance
}

func (s *Slave) Run() {
	log.Println("Running slave on port", s.port)

	go func() {
		http.HandleFunc(lib.TasksCreatePath, taskHandler)
		http.HandleFunc(lib.JobsKillPath, makeKillJobHandler(s.killChan))
		err := http.ListenAndServe(":"+s.port, nil)
		if err != nil {
			log.Panicln("[Main Loop] listener failed", err)
		}
	}()

	s.heartbeatTicker = time.NewTicker(lib.HeartbeatInterval)
	for {
		select {
		case <-s.heartbeatTicker.C:
			s.sendHeartbeat()

		case t := <-s.addTaskChan:
			taskJSON, _ := t.ToJSON()
			log.Println("[Main Loop]", "Add task:", string(taskJSON))
			task := &task{
				Task:         t,
				Status:       lib.Running,
				progressChan: make(chan chan []byte),
				killChan:     make(chan bool),
			}
			if s.addTask(task) {
				go execute(task, s.successChan, s.failChan)
			}

		case msg := <-s.successChan:
			log.Println("[Main Loop]", "SuccessChan message:", msg)
			s.passwordFound(msg.taskID, msg.password)

		case msg := <-s.failChan:
			log.Println("[Main Loop]", "FailChan message:", msg)
			s.passwordNotFound(msg.taskID)

		case jobID := <-s.killChan:
			for i := range s.tasks {
				if s.tasks[i].JobID == jobID && s.tasks[i].Status == lib.Running {
					s.tasks[i].killChan <- true
					s.passwordNotFound(s.tasks[i].ID)
					log.Println("[Main loop] killed job ", jobID, "task", s.tasks[i].ID)
				}
			}
		}

	}
}

func taskHandler(w http.ResponseWriter, r *http.Request) {
	var t lib.Task
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&t); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Println("[taskHandler] sending new task into channel", t.ID)
	slaveInstance.addTaskChan <- t
}

func makeKillJobHandler(c chan int) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		jobID := r.FormValue("jobid")
		res, err := strconv.ParseInt(jobID, 10, 64)
		if err != nil {
			log.Panicln("[killJobHandler] failed to parse", jobID, err)
		}
		log.Println("[jobsHandler] sending kill request into channel", res)
		c <- int(res)
	}
}

func (s *Slave) addTask(task *task) bool {
	if s.taskWithID(task.ID) == nil {
		s.tasks = append(s.tasks, task)
		return true
	}
	log.Println("[ERROR]", "Task with id", task.ID, "already exists. Request discarded.")
	return false
}

func (s *Slave) passwordFound(id int, password string) {
	log.Println("[ Task", id, "]", "Found password:", password)
	t := s.taskWithID(id)
	if t != nil {
		t.Status = lib.PasswordFound
		t.Password = password
	} else {
		log.Println("[ERROR]", "Id not found in Taskstatus")
	}
}

func (s *Slave) passwordNotFound(id int) {
	log.Println("[ Task", id, "]", "Password not found")
	t := s.taskWithID(id)
	if t != nil {
		t.Status = lib.PasswordNotFound
	} else {
		log.Println("[ERROR]", "Id not found in Taskstatus")
	}
}

func (s *Slave) taskWithID(id int) *task {
	for i, t := range s.tasks {
		if t.ID == id {
			return s.tasks[i]
		}
	}
	return nil
}

func (s *Slave) removeTaskWithID(id int) {
	for i := range s.tasks {
		if s.tasks[i].ID == id {
			s.tasks = append(s.tasks[:i], s.tasks[i+1:]...)
			log.Println("[removeTaskWithID] removed task with id", id)
			return
		}
	}
	log.Panicln("[removeTaskWithID] failed to remove task with id", id)
}

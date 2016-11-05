package slave

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/aaronang/cong-the-ripper/lib"
)

var slaveInstance Slave

type task struct {
	lib.Task
	Status       lib.Status
	Password     string
	progressChan chan chan string
}

type Slave struct {
	id              string
	port            string
	masterIp        string
	masterPort      string
	successChan     chan CrackerSuccess
	failChan        chan CrackerFail
	addTaskChan     chan lib.Task
	heartbeatTicker *time.Ticker
	tasks           []*task
}

func Init(instanceId, port, masterIp, masterPort string) *Slave {
	slaveInstance = Slave{
		id:              instanceId,
		port:            port,
		masterIp:        masterIp,
		masterPort:      masterPort,
		successChan:     make(chan CrackerSuccess),
		failChan:        make(chan CrackerFail),
		addTaskChan:     make(chan lib.Task),
		heartbeatTicker: nil,
	}
	return &slaveInstance
}

func (s *Slave) Run() {
	log.Println("Running slave on port", s.port)

	go func() {
		http.HandleFunc(lib.TasksCreatePath, taskHandler)
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
				progressChan: make(chan chan string),
			}
			if s.addTask(task) {
				go Execute(task, s.successChan, s.failChan)
			}

		case msg := <-s.successChan:
			log.Println("[Main Loop]", "SuccessChan message:", msg)
			s.passwordFound(msg.taskID, msg.password)

		case msg := <-s.failChan:
			log.Println("[Main Loop]", "FailChan message:", msg)
			s.passwordNotFound(msg.taskID)
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
	slaveInstance.addTaskChan <- t
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

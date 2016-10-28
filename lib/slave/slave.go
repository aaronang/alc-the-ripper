package slave

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/aaronang/cong-the-ripper/lib"
)

var slaveInstance Slave

type Slave struct {
	port        string
	masterIp    string
	masterPort  string
	heartbeat   lib.Heartbeat
	successChan chan CrackerSuccess
	failChan    chan CrackerFail
	addTaskChan chan lib.Task
}

func Init(instanceId, port, masterIp, masterPort string) *Slave {
	heartbeat := lib.Heartbeat{
		SlaveId: instanceId,
	}

	slaveInstance = Slave{
		port:        port,
		masterIp:    masterIp,
		masterPort:  masterPort,
		heartbeat:   heartbeat,
		successChan: make(chan CrackerSuccess),
		failChan:    make(chan CrackerFail),
		addTaskChan: make(chan lib.Task),
	}
	return &slaveInstance
}

func (s *Slave) Run() {
	http.HandleFunc(lib.TasksCreatePath, taskHandler)
	go http.ListenAndServe(":"+s.port, nil)
	log.Println("Running slave on port", s.port)
	go s.HeartbeatSender()

	for {
		select {
		case task := <-s.addTaskChan:
			s.addTask(task)
			go Execute(task, s.successChan, s.failChan)

		case msg := <-s.successChan:
			s.passwordFound(msg.taskID, msg.password)

		case msg := <-s.failChan:
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

func (s *Slave) addTask(task lib.Task) {
	taskStatus := lib.TaskStatus{
		Id:       task.ID,
		JobId:    task.JobID,
		Status:   lib.Running,
		Progress: task.Start,
	}
	s.heartbeat.TaskStatus = append(s.heartbeat.TaskStatus, taskStatus)
}

func (s *Slave) passwordFound(id int, password string) {
	log.Println("Found password: " + password)
	ts := s.taskStatusWithId(id)
	if ts != nil {
		ts.Status = lib.PasswordFound
		ts.Password = password
	} else {
		log.Println("ERROR:", "Id not found in Taskstatus")
	}
}

func (s *Slave) passwordNotFound(id int) {
	log.Println("Password not found")
	ts := s.taskStatusWithId(id)
	if ts != nil {
		ts.Status = lib.PasswordNotFound
	} else {
		log.Println("ERROR:", "Id not found in Taskstatus")
	}
}

func (s *Slave) taskStatusWithId(id int) *lib.TaskStatus {
	for i, ts := range s.heartbeat.TaskStatus {
		if ts.Id == id {
			return &s.heartbeat.TaskStatus[i]
		}
	}
	return nil
}

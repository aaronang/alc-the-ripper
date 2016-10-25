package slave

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aaronang/cong-the-ripper/lib"
)

var slaveInstance Slave

type Slave struct {
	heartbeat   lib.Heartbeat
	successChan chan CrackerSuccess
	failChan    chan CrackerFail
	addTaskChan chan lib.Task
}

func Init(instanceId string) *Slave {
	heartbeat := lib.Heartbeat{
		SlaveId: instanceId,
	}

	slaveInstance = Slave{
		heartbeat:   heartbeat,
		successChan: make(chan CrackerSuccess),
		failChan:    make(chan CrackerFail),
		addTaskChan: make(chan lib.Task),
	}
	return &slaveInstance
}

func (s *Slave) Run() {
	http.HandleFunc(lib.TasksCreatePath, taskHandler)
	go http.ListenAndServe(lib.Port, nil)

	for {
		select {
		case task := <-s.addTaskChan:
			s.addTask(task)
			go Execute(task, s.successChan, s.failChan)

		case msg := <-s.successChan:
			s.password_found(msg.taskID, msg.password)

		case msg := <-s.failChan:
			s.password_not_found(msg.taskID)
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

func (s *Slave) password_found(Id int, password string) {
	fmt.Println("Found password: " + password)
	ts := s.taskStatusWithId(Id)
	if ts != nil {
		ts.Status = lib.PasswordFound
		ts.Password = password
	} else {
		fmt.Println("ERROR:", "Id not found in Taskstatus")
	}
}

func (s *Slave) password_not_found(Id int) {
	fmt.Println("Password not found")
	ts := s.taskStatusWithId(Id)
	if ts != nil {
		ts.Status = lib.PasswordNotFound
	} else {
		fmt.Println("ERROR:", "Id not found in Taskstatus")
	}
}

func (s *Slave) taskStatusWithId(Id int) *lib.TaskStatus {
	for i, ts := range s.heartbeat.TaskStatus {
		if ts.Id == Id {
			return &s.heartbeat.TaskStatus[i]
		}
	}
	return nil
}

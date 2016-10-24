package slave

import (
	"encoding/json"
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
			go s.Execute(task)

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

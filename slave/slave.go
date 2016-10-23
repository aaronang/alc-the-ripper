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
	heartbeat := lib.Heartbeat {
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
		case task := <- s.addTaskChan:
			fmt.Println("add task")
			taskStatus := lib.TaskStatus {
				Id: task.ID,
				JobId: task.JobID,
				Done: false,
				Progress: task.Start,
			}
			s.heartbeat.TaskStatus = append(s.heartbeat.TaskStatus, taskStatus)
			fmt.Println(s.heartbeat.TaskStatus)

		case msg := <- s.successChan:
			fmt.Println("Found password: " + msg.Password)

		case <- s.failChan:
			fmt.Println("Password not found")
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
	fmt.Println(t)
	go Execute(t, &slaveInstance)
}

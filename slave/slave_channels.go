package slave

import (
	"fmt"

	"github.com/aaronang/cong-the-ripper/lib"
)

func addTask(task lib.Task, s *Slave) {
	taskStatus := lib.TaskStatus{
		Id:       task.ID,
		JobId:    task.JobID,
		Done:     false,
		Progress: task.Start,
	}
	s.heartbeat.TaskStatus = append(s.heartbeat.TaskStatus, taskStatus)
}

func password_found(Id int, password string, s *Slave) {
	fmt.Println("Found password: " + password)
	ts := taskStatusWithId(Id, s)
	if ts != nil {
		ts.Done = true
		ts.Password = password
	}
}

func password_not_found(Id int, s *Slave) {
	fmt.Println("Password not found")
	ts := taskStatusWithId(Id, s)
	if ts != nil {
		ts.Done = true
	}
}

func taskStatusWithId(Id int, s *Slave) *lib.TaskStatus {
	for i, ts := range s.heartbeat.TaskStatus {
		if ts.Id == Id {
			return &s.heartbeat.TaskStatus[i]
		}
	}
	return nil
}

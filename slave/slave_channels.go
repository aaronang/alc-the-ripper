package slave

import (
	"fmt"

	"github.com/aaronang/cong-the-ripper/lib"
)

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

package slave

import (
	"bytes"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/aaronang/cong-the-ripper/lib"
)

func (s *Slave) sendHeartbeat() {
	log.Println("[Heartbeat] Sending...")
	heartbeat := s.generateHeartbeat()
	_, err := s.postHeartbeat(heartbeat)

	if err != nil {
		log.Println("[Heartbeat] Delivery failed")
	}
}

func (s *Slave) generateHeartbeat() lib.Heartbeat {
	heartbeat := lib.Heartbeat{
		SlaveId: s.id,
	}
	for i, task := range s.tasks {
		var progress []byte
		if task.Status == lib.Running {
			c := make(chan []byte)
			s.tasks[i].progressChan <- c
			select {
			case progress = <-c:
			case <-time.After(time.Second * 1): // This could accually happen for a legitimate reason (much hash interations, slow hashing algorithms), but for now see what happens
				log.Println("[ERROR]", "progressChan did not respond for task", task.ID)
			}
		}
		heartbeat.TaskStatus = append(heartbeat.TaskStatus, lib.TaskStatus{
			Id:       task.ID,
			JobId:    task.JobID,
			Status:   task.Status,
			Password: task.Password,
			Progress: progress, // Can be empty
		})
	}
	return heartbeat
}

func (s *Slave) postHeartbeat(heartbeat lib.Heartbeat) (*http.Response, error) {
	url := lib.Protocol + net.JoinHostPort(s.masterIp, s.masterPort) + lib.HeartbeatPath
	body, err := heartbeat.ToJSON()
	if err != nil {
		panic(err)
	}

	timeout := time.Duration(lib.RequestTimeout)
	client := http.Client{
		Timeout: timeout,
	}
	return client.Post(url, lib.BodyType, bytes.NewBuffer(body))
}

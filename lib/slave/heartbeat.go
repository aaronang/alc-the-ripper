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
	_, err := postHeartbeat(s)

	if err == nil {
		var taskStatusses []lib.TaskStatus
		for _, ts := range s.heartbeat.TaskStatus {
			if ts.Status == lib.Running {
				taskStatusses = append(taskStatusses, ts)
			}
		}
		s.heartbeat.TaskStatus = taskStatusses
	} else {
		log.Println("[Heartbeat] Delivery failed")
	}
}

func postHeartbeat(s *Slave) (*http.Response, error) {
	url := lib.Protocol + net.JoinHostPort(s.masterIp, s.masterPort) + lib.HeartbeatPath
	body, err := s.heartbeat.ToJSON()
	if err != nil {
		panic(err)
	}

	timeout := time.Duration(lib.RequestTimeout)
	client := http.Client{
		Timeout: timeout,
	}
	return client.Post(url, lib.BodyType, bytes.NewBuffer(body))
}

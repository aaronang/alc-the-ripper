package slave

import (
	"bytes"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/aaronang/cong-the-ripper/lib"
)

func (s *Slave) HeartbeatSender() {
	for {
		select {
		case <-time.After(time.Second * 5):
			log.Println("Heartbeat...")
			_, err := SendHeartbeat(s)
			if err != nil {
				log.Println("Heartbeat to master failed.")
			}
		}
	}
}

func SendHeartbeat(s *Slave) (*http.Response, error) {
	url := lib.Protocol + net.JoinHostPort(s.masterIp, s.masterPort) + lib.HeartbeatPath
	body, err := s.heartbeat.ToJSON()
	if err != nil {
		panic(err)
	}

	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	return client.Post(url, lib.BodyType, bytes.NewBuffer(body))
}

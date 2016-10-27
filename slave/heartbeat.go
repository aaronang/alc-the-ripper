package slave

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/aaronang/cong-the-ripper/lib"
)

func (s *Slave) HeartbeatSender() {
	for {
		select {
		case <-time.After(time.Second * 1):
			fmt.Println("Heartbeat...")
		}
	}
}

func SendHeartbeat(s *Slave) (*http.Response, error) {
	url := lib.Protocol + "localhost" + s.port + lib.HeartbeatPath
	body, err := s.heartbeat.ToJSON()
	if err != nil {
		panic(err)
	}
	return http.Post(url, lib.BodyType, bytes.NewBuffer(body))
}

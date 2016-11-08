package slave

import (
	b64 "encoding/base64"
	"testing"
	"time"

	"github.com/aaronang/cong-the-ripper/lib"
)

func setupSlaveTask() (*Slave, *task) {
	slave := Init("8080", "127.0.0.1", "3000")
	slave.successChan = make(chan CrackerSuccess)
	slave.failChan = make(chan CrackerFail)

	digestBytes, _ := b64.StdEncoding.DecodeString("WTpSrbQAR8IMSK9uMoOQEXfKy+2FojN8yEz+T1n21uE=") //cong

	job := lib.Job{
		Salt:      []byte("salty"),
		Digest:    digestBytes,
		KeyLen:    22,
		Iter:      6400,
		Alphabet:  lib.AlphaLower,
		Algorithm: lib.PBKDF2,
	}

	task := &task{
		Task: lib.Task{
			Job:     job,
			JobID:   1,
			ID:      1,
			Start:   []byte("anoc"),
			TaskLen: 26,
		},
		Status:       lib.Running,
		progressChan: make(chan chan []byte),
	}
	return slave, task
}

func TestHit(t *testing.T) {
	slave, task := setupSlaveTask()
	go Execute(task, slave.successChan, slave.failChan)
	select {
	case <-time.After(time.Second * 10):
	case <-slave.failChan:
		t.Fail()
	case <-slave.successChan:
	}
}

func TestMiss(t *testing.T) {
	slave, task := setupSlaveTask()
	task.TaskLen = 1
	go Execute(task, slave.successChan, slave.failChan)
	select {
	case <-slave.successChan:
		t.Fail()
	case <-slave.failChan:
	}
}

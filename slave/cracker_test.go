package slave

import (
	b64 "encoding/base64"
	"testing"
	"time"

	"github.com/aaronang/cong-the-ripper/lib"
)

var task lib.Task
var slave *Slave

func Setup() {
	slave = Init("instance.EC2.cong1")
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

	task = lib.Task{
		Job:     job,
		JobID:   1,
		ID:      1,
		Start:   []byte("anoc"),
		TaskLen: 26,
	}
}

func TestHit(t *testing.T) {
	Setup()
	go Execute(task, slave.successChan, slave.failChan)
	select {
	case <-time.After(time.Second * 10):
	case <-slave.failChan:
		t.Fail()
	case <-slave.successChan:
	}
}

func TestMiss(t *testing.T) {
	Setup()
	task.TaskLen = 1
	go Execute(task, slave.successChan, slave.failChan)
	select {
	case <-slave.successChan:
		t.Fail()
	case <-slave.failChan:
	}
}

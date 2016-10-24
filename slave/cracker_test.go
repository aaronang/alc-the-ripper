package slave

import (
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

	job := lib.Job{
		Salt:      []byte("salty"),
		Digest:    "WTpSrbQAR8IMSK9uMoOQEXfKy+2FojN8yEz+T1n21uE=", // cong
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
	go Execute(task, slave)
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
	go Execute(task, slave)
	select {
	case <-slave.successChan:
		t.Fail()
	case <-slave.failChan:
	}
}

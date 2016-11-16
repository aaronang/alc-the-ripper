package slave

import (
	"bytes"
	b64 "encoding/base64"
	"testing"
	"time"

	"github.com/aaronang/cong-the-ripper/lib"
	"github.com/aaronang/cong-the-ripper/lib/slave/brutedict"
)

func setupSlaveTask() (*Slave, *task) {
	slave := Init("8080", "127.0.0.1", "3000")
	slave.successChan = make(chan CrackerSuccess)
	slave.failChan = make(chan CrackerFail)

	digest, _ := b64.StdEncoding.DecodeString("Oamol38L3PkwaQ2SR3AHIh/eyzh6Ltvv0bJiyDk1l4w=") //afpl
	salt, _ := b64.StdEncoding.DecodeString("lkw=")

	job := lib.Job{
		Salt:      salt,
		Digest:    digest,
		KeyLen:    4,
		Iter:      2000,
		Alphabet:  lib.AlphaLower,
		Algorithm: lib.PBKDF2,
	}

	task := &task{
		Task: lib.Task{
			Job:     job,
			JobID:   1,
			ID:      1,
			Start:   []byte("apfa"),
			TaskLen: 142000,
		},
		Status:       lib.Running,
		progressChan: make(chan chan []byte),
	}

	return slave, task
}

func TestHit(t *testing.T) {
	slave, task := setupSlaveTask()
	go execute(task, slave.successChan, slave.failChan)
	select {
	case <-time.After(time.Second * 10):
		t.Fail()
	case <-slave.failChan:
		t.Fail()
	case <-slave.successChan:
	}
}

func TestProgressHit(t *testing.T) {
	_, task := setupSlaveTask()
	task.Progress = []byte("dcba")
	bd := brutedict.New(&task.Task)
	bd.Next()
	if bytes.Compare(bd.Next(), []byte("abce")) != 0 {
		t.Fail()
	}
	if bytes.Compare(bd.Next(), []byte("abcf")) != 0 {
		t.Fail()
	}
}

func TestMiss(t *testing.T) {
	slave, task := setupSlaveTask()
	task.TaskLen = 1
	go execute(task, slave.successChan, slave.failChan)
	select {
	case <-slave.successChan:
		t.Fail()
	case <-slave.failChan:
	}
}

func TestHashRate(t *testing.T) {
	slave, task := setupSlaveTask()
	task.Digest = []byte("wrongdigest")
	task.TaskLen = 5000 // TaskLen * Iter (2000) = 10 million
	go execute(task, slave.successChan, slave.failChan)
	select {
	case <-slave.successChan:
		t.Fail()
	case <-slave.failChan:
	}
}

package cracker

import (
	"testing"

	"github.com/aaronang/cong-the-ripper/lib"
)

var task lib.Task

func Setup() {
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
	Execute(task)
}

func TestMiss(t *testing.T) {
	Setup()
	task.TaskLen = 1
	Execute(task)
}

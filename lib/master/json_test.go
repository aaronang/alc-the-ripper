package master

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/aaronang/cong-the-ripper/lib"
)

func TestCreateStatusJSON(t *testing.T) {
	lj1 := lib.Job{
		Salt:      []byte("salt"),
		Digest:    []byte("digest"),
		KeyLen:    4,
		Iter:      22,
		Alphabet:  lib.AlphaLower,
		Algorithm: lib.PBKDF2,
	}

	t1 := &lib.Task{
		Job:     lj1,
		JobID:   1,
		ID:      1,
		Start:   []byte("aaaa"),
		TaskLen: 22,
	}

	t2 := &lib.Task{
		Job:     lj1,
		JobID:   1,
		ID:      2,
		Start:   []byte("waa"),
		TaskLen: 22,
	}

	j1 := &job{
		Job:          lj1,
		id:           1,
		tasks:        []*lib.Task{t1, t2},
		runningTasks: 2,
		maxTasks:     2,
	}

	s1 := slave{
		tasks:    []*lib.Task{t1, t2},
		maxSlots: 3,
	}

	m := &Master{
		instances: map[string]slave{
			"52.51.156.198": s1,
		},
		jobs: map[int]*job{
			1: j1,
		},
	}

	status := createStatusJSON(m)

	js, err := json.MarshalIndent(status, "", "\t")
	if err != nil {
		t.Error("Unable to marshal status summary", err)
	}
	log.Println(string(js))
}

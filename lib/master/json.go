package master

import (
	"github.com/aaronang/cong-the-ripper/lib"
	"time"
)

type StatusJSON struct {
	Slaves        []SlaveJSON `json:"slaves"`
	Jobs          []JobJSON   `json:"jobs"`
	CompletedJobs []JobJSON   `json:"completedJobs"`
}

type SlaveJSON struct {
	IP       string     `json:"ip"`
	MaxSlots int        `json:"maxSlots"`
	Tasks    []TaskJSON `json:"tasks"`
}

type TaskJSON struct {
	ID       int    `json:"id"`
	JobID    int    `json:"jobId"`
	Start    []byte `json:"start"`
	TaskLen  int    `json:"taskLen"`
	Progress []byte `json:"progress"`
}

type JobJSON struct {
	ID         int           `json:"id"`
	Salt       []byte        `json:"salt"`
	Digest     []byte        `json:"digest"`
	KeyLen     int           `json:"keyLen"`
	Iter       int           `json:"iter"`
	Alphabet   lib.Alphabet  `json:"alphabet"`
	Algorithm  lib.Algorithm `json:"algorithm"`
	Tasks      []TaskJSON    `json:"tasks"`
	StartTime  time.Time     `json:"startTime"`
	FinishTime time.Time     `json:"finishTime"`
	Password   string        `json:"password"`
}

func createStatusJSON(m *Master) StatusJSON {
	return StatusJSON{
		Slaves:        createSlavesJSON(m.instances),
		Jobs:          createJobsJSON(m.jobs),
		CompletedJobs: createJobsJSON(m.completedJobs),
	}
}

func createSlavesJSON(ss map[string]slave) []SlaveJSON {
	var slaves []SlaveJSON
	for ip, s := range ss {
		slave := SlaveJSON{
			IP:       ip,
			MaxSlots: s.maxSlots,
			Tasks:    createTasksJSON(s.tasks),
		}
		slaves = append(slaves, slave)
	}
	return slaves
}

func createTasksJSON(ts []*lib.Task) []TaskJSON {
	var tasks []TaskJSON
	for _, t := range ts {
		task := TaskJSON{
			ID:       t.ID,
			JobID:    t.JobID,
			Start:    t.Start,
			TaskLen:  t.TaskLen,
			Progress: t.Progress,
		}
		tasks = append(tasks, task)
	}
	return tasks
}

func createJobsJSON(js map[int]*job) []JobJSON {
	var jobs []JobJSON
	for _, j := range js {
		job := JobJSON{
			ID:         j.id,
			Salt:       j.Salt,
			Digest:     j.Digest,
			KeyLen:     j.KeyLen,
			Iter:       j.Iter,
			Alphabet:   j.Alphabet,
			Algorithm:  j.Algorithm,
			Tasks:      createTasksJSON(j.tasks),
			StartTime:  j.startTime,
			FinishTime: j.finishTime,
			Password:   j.password,
		}
		jobs = append(jobs, job)
	}
	return jobs
}

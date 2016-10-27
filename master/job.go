package master

import (
	"log"
	"math/big"

	"github.com/aaronang/cong-the-ripper/lib"
)

// NOTE: consider using math/big for some of the operations in this package

type job struct {
	lib.Job
	id           int
	tasks        []*lib.Task
	runningTasks int
	maxTasks     int
}

func (j *job) reachedMaxTasks() bool {
	return j.runningTasks < j.maxTasks
}

func (j *job) increaseRunningTasks() {
	if j.runningTasks >= len(j.tasks) || j.runningTasks >= j.maxTasks {
		log.Fatalln("Trying to run more tasks than possible or allowed.")
	}
	j.runningTasks = j.runningTasks + 1
}

func (j *job) decreaseRunningTasks() {
	if j.runningTasks <= 0 {
		log.Fatalln("Running tasks can never be lower than zero.")
	}
	j.runningTasks = j.runningTasks - 1
}

// SplitJob attempts to split a cracking job into equal sized tasks regardless of the job
// the taskSize represents the number of brute force iterations
func SplitJob(job *job, taskSize int64) []lib.Task {
	var tasks []lib.Task
	cands, lens := chunkCandidates(job.Alphabet, job.KeyLen, taskSize)
	for i := range cands {
		tasks = append(tasks, lib.Task{
			Job:     job.Job,
			JobID:   job.id,
			ID:      i,
			Start:   cands[i],
			TaskLen: lens[i]})
	}
	return tasks
}

// chunkCandidates takes a character set and the required length l and splits to chunks of size n
func chunkCandidates(alph lib.Alphabet, l int, n int64) ([][]byte, []int64) {
	cand := alph.InitialCandidate(l)
	var cands [][]byte
	var lens []int64
	for {
		newCand, overflow := nthCandidateFrom(alph, n, cand)
		cands = append(cands, cand)
		if overflow {
			lens = append(lens, countUntilFinal(alph, cand))
			break
		} else {
			lens = append(lens, n)
		}
		cand = newCand
	}
	return cands, lens
}

// nthCandidateFrom computes the n th candidate password from inp
func nthCandidateFrom(alph lib.Alphabet, n int64, inp []byte) ([]byte, bool) {
	l := len(inp)
	z := lib.BytesToBigInt(alph, inp)
	z.Add(z, big.NewInt(n))
	return lib.BigIntToBytes(alph, z, l)
}

// countUntilFinal counts the number of iterations until the final candidate starting from cand
func countUntilFinal(alph lib.Alphabet, cand []byte) int64 {
	c := lib.BytesToBigInt(alph, cand)
	f := lib.BytesToBigInt(alph, alph.FinalCandidate(len(cand)))
	diff := f.Sub(f, c)
	return diff.Int64()
}

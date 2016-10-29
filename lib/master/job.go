package master

import (
	"log"
	"math/big"

	"github.com/aaronang/cong-the-ripper/lib"
)

type job struct {
	lib.Job
	id           int
	tasks        []*lib.Task
	runningTasks int
	maxTasks     int
}

type task struct {
	lib.Task
	done bool
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

// splitJob attempts to split a cracking job into equal sized tasks regardless of the job
// the taskSize represents the number of brute force iterations
func (j *job) splitJob(taskSize int) {
	log.Printf("[splitJob] splitting job %v to size %v\n", j.id, taskSize)
	if taskSize < j.Iter {
		log.Panicln("[splitJob] taskSize cannot be lower than job.Iter")
	}

	// adjust taskSize depending on the PBKDF2 rounds
	actualTaskSize := taskSize / j.Iter

	var tasks []*lib.Task
	cands, lens := chunkCandidates(j.Alphabet, j.KeyLen, actualTaskSize)
	for i := range cands {
		tasks = append(tasks, &lib.Task{
			Job:     j.Job,
			JobID:   j.id,
			ID:      i,
			Start:   cands[i],
			TaskLen: lens[i]})
	}
	j.tasks = tasks
	log.Printf("[splitJob] Done, job %v has %v tasks\n", j.id, len(j.tasks))
}

// chunkCandidates takes a character set and the required length l and splits to chunks of size n
func chunkCandidates(alph lib.Alphabet, l int, n int) ([][]byte, []int) {
	cand := alph.InitialCandidate(l)
	var cands [][]byte
	var lens []int
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
func nthCandidateFrom(alph lib.Alphabet, n int, inp []byte) ([]byte, bool) {
	l := len(inp)
	z := lib.BytesToBigInt(alph, inp)
	z.Add(z, big.NewInt(int64(n)))
	return lib.BigIntToBytes(alph, z, l)
}

// countUntilFinal counts the number of iterations until the final candidate starting from cand
func countUntilFinal(alph lib.Alphabet, cand []byte) int {
	c := lib.BytesToBigInt(alph, cand)
	f := lib.BytesToBigInt(alph, alph.FinalCandidate(len(cand)))
	diff := f.Sub(f, c)
	return int(diff.Int64())
}

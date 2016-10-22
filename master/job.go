package master

import (
	"github.com/aaronang/cong-the-ripper/lib"
)

// NOTE: consider using math/big for some of the operations in this package

type job struct {
	lib.Job
	id    int
	tasks []lib.Task
}

// SplitJob attempts to split a cracking job into equal sized tasks regardless of the job
// the taskSize represents the number of brute force iterations
func SplitJob(job *job, taskSize int) []lib.Task {
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
func chunkCandidates(alph lib.Alphabet, l, n int) ([][]byte, []int) {
	cand := alph.InitialCandidate(l)
	var cands [][]byte
	var lens []int
	for {
		newCand, carry := nthCandidateFrom(alph, n, cand)
		if carry == 0 {
			lens = append(lens, n)
			cands = append(cands, cand)
		} else {
			// this part is not efficient, if the worker has the ability to run
			// the task until it reaches the end without sacrifising performance
			// then we can use -1 to indicate those types of tasks
			lens = append(lens, countUntilFinal(alph, cand))
			cands = append(cands, cand)
			break
		}
		cand = newCand
	}
	return cands, lens
}

// nthCandidateFrom computes the n th candidate password from inp
func nthCandidateFrom(alph lib.Alphabet, n int, inp []byte) ([]byte, int) {
	base := len(lib.Alphabets[alph])
	res, carry := lib.AddToIntSlice(base, n, lib.BytesToIntSlice(alph, inp))
	return lib.IntSliceToBytes(alph, res), carry
}

// countUntilFinal counts the number of iterations until the final candidate starting from cand
// not so efficient
// we can use binary search to improve the performance
func countUntilFinal(alph lib.Alphabet, cand []byte) int {
	f := lib.BytesToIntSlice(alph, alph.FinalCandidate(len(cand)))
	cand2 := lib.BytesToIntSlice(alph, cand)
	base := len(lib.Alphabets[alph])
	i := 0
	for !lib.TestEqInts(f, cand2) {
		cand2, _ = lib.AddToIntSlice(base, 1, cand2)
		i++
	}
	return i
}

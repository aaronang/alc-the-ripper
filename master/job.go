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
	cands, lens := chunkCharSet(job.CharSet, job.KeyLen, taskSize)
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

// chunkCharSet takes a character set and the required length l and splits to chunks of size n
func chunkCharSet(charset lib.CharSet, l, n int) ([][]byte, []int) {
	cand := initialCandidate(charset, l)
	var cands [][]byte
	var ints []int
	for {
		newComb, carry := nthCandidateFrom(charset, n, cand)
		if carry == 0 {
			ints = append(ints, n)
			cands = append(cands, cand)
		} else {
			// this part is not efficient, if the worker has the ability to run
			// the task until it reaches the end without sacrifising performance
			// then we can use -1 to indicate those types of tasks
			ints = append(ints, countUntilFinal(charset, cand))
			cands = append(cands, cand)
			break
		}
		cand = newComb
	}
	return cands, ints
}

// nthCandidateFrom computes the n th candidate password from inp
func nthCandidateFrom(charset lib.CharSet, n int, inp []byte) ([]byte, int) {
	base := len(lib.CharSetSlice[charset])
	res, carry := lib.AddToIntSlice(base, n, lib.BytesToIntSlice(charset, inp))
	return lib.IntSliceToBytes(charset, res), carry
}

// countUntilFinal counts the number of iterations until the final candidate starting from cand
// not so efficient
// we can use binary search to improve the performance
func countUntilFinal(charset lib.CharSet, cand []byte) int {
	f := lib.BytesToIntSlice(charset, finalCandidate(charset, len(cand)))
	combi := lib.BytesToIntSlice(charset, cand)
	base := len(lib.CharSetSlice[charset])
	i := 0
	for !lib.TestEqInts(f, combi) {
		combi, _ = lib.AddToIntSlice(base, 1, combi)
		i++
	}
	return i
}

func initialCandidate(charset lib.CharSet, l int) []byte {
	return replicateAt(charset, l, 0)
}

func finalCandidate(charset lib.CharSet, l int) []byte {
	return replicateAt(charset, l, len(lib.CharSetSlice[charset])-1)
}

func replicateAt(charset lib.CharSet, l int, idx int) []byte {
	v := lib.CharSetSlice[charset][idx]
	res := make([]byte, l)
	for i := range res {
		res[i] = v
	}
	return res
}

package master

import (
	"bytes"

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
	res, carry := addToIntSlice(base, n, bytesToIntSlice(charset, inp))
	return intSliceToBytes(charset, res), carry
}

// countUntilFinal counts the number of iterations until the final candidate starting from cand
// not so efficient
// we can use binary search to improve the performance
func countUntilFinal(charset lib.CharSet, cand []byte) int {
	f := bytesToIntSlice(charset, finalCandidate(charset, len(cand)))
	combi := bytesToIntSlice(charset, cand)
	base := len(lib.CharSetSlice[charset])
	i := 0
	for !testEq(f, combi) {
		combi, _ = addToIntSlice(base, 1, combi)
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

func bytesToIntSlice(charset lib.CharSet, inp []byte) []int {
	res := make([]int, len(inp))
	for i, b := range inp {
		// probably not efficient, but the character sets are small so it's negligible
		x := bytes.IndexByte(lib.CharSetSlice[charset], b)
		if x < 0 {
			panic("Invalid characters!")
		}
		res[i] = x
	}
	return res
}

func addToIntSlice(base, v int, inp []int) ([]int, int) {
	for i := range inp {
		r := v % base
		tmp := inp[i] + r
		if tmp < base {
			inp[i] = tmp
			v = v / base
		} else {
			inp[i] = tmp % base
			v = v/base + 1
		}
	}
	return inp, v // v is the carry
}

func intSliceToBytes(charset lib.CharSet, inp []int) []byte {
	res := make([]byte, len(inp))
	for i, v := range inp {
		res[i] = lib.CharSetSlice[charset][v]
	}
	return res
}

func testEq(a, b []int) bool {

	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

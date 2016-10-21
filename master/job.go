package master

import (
	"bytes"

	"github.com/aaronang/cong-the-ripper/lib"
)

var charSetSlice [][]byte

func init() {
	nums := "0123456789"
	alphaLower := "abcdefghijklmnopqrstuvwxyz"
	alphaMixed := alphaLower + "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	alphaNumLower := nums + alphaLower
	charSetSlice = [][]byte{
		[]byte(nums),
		[]byte(alphaLower),
		[]byte(alphaMixed),
		[]byte(alphaNumLower)}
}

type job struct {
	lib.Job
	id int
	// Tasks []lib.Task
}

// SplitJob attempts to split a cracking job into equal sized tasks
func SplitJob(job *job, size int) []lib.Task {
	var tasks []lib.Task
	combs, lens := chunkCharSet(job.CharSet, job.KeyLen, size)
	for i := range combs {
		tasks = append(tasks, lib.Task{
			Job:     job.Job,
			ID:      i,
			Start:   combs[i],
			TaskLen: lens[i]})
	}
	return tasks
}

func chunkCharSet(charset lib.CharSet, l, n int) ([][]byte, []int) {
	comb := initialCombination(charset, l)
	var combs [][]byte
	var ints []int
	for {
		newComb, carry := nextCombination(charset, n, comb)
		if carry == 0 {
			ints = append(ints, n)
			combs = append(combs, comb)
		} else {
			// this part is not efficient, if the worker has the ability to run
			// the task until it reaches the end without sacrifising performance
			// then we can use -1 to indicate those types of tasks
			ints = append(ints, countUntilFinal(charset, comb))
			combs = append(combs, comb)
			break
		}
		comb = newComb
	}
	return combs, ints
}

func nextCombination(charset lib.CharSet, v int, inp []byte) ([]byte, int) {
	base := len(charSetSlice[charset])
	res, carry := addToIntSlice(base, v, bytesToIntSlice(charset, inp))
	return intSliceToBytes(charset, res), carry
}

// not so efficient
// we can use binary search to improve the performance
func countUntilFinal(charset lib.CharSet, comb []byte) int {
	f := bytesToIntSlice(charset, finalCombination(charset, len(comb)))
	combi := bytesToIntSlice(charset, comb)
	base := len(charSetSlice[charset])
	i := 0
	for !testEq(f, combi) {
		combi, _ = addToIntSlice(base, 1, combi)
		i++
	}
	return i
}

func initialCombination(charset lib.CharSet, l int) []byte {
	return replicateAt(charset, l, 0)
}

func finalCombination(charset lib.CharSet, l int) []byte {
	return replicateAt(charset, l, len(charSetSlice[charset])-1)
}

func replicateAt(charset lib.CharSet, l int, idx int) []byte {
	v := charSetSlice[charset][idx]
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
		x := bytes.IndexByte(charSetSlice[charset], b)
		if x < 0 {
			panic("Invalid characters!")
		}
		res[i] = x
	}
	return res
}

func addToIntSlice(base, v int, inp []int) (result []int, carry int) {
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
	return inp, v
}

func intSliceToBytes(charset lib.CharSet, inp []int) []byte {
	res := make([]byte, len(inp))
	for i, v := range inp {
		res[i] = charSetSlice[charset][v]
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

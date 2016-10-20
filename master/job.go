package master

import (
	"bytes"

	"github.com/aaronang/cong-the-ripper/lib"
)

var charSetSlice [][]byte

func init() {
	nums := []byte("0123456789")
	alphaLower := []byte("abcdefghijklmnopqrstuvwxyz")
	alphaMixed := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	alphaNumLower := []byte("0123456789abcdefghijklmnopqrstuvwxyz")
	charSetSlice = [][]byte{nums, alphaLower, alphaMixed, alphaNumLower}
}

type job struct {
	lib.Job
	id int
	// Tasks []lib.Task
}

// func SplitJob(job *job, size int) []lib.Task {
// 	var tasks []lib.Task
// 	comb := initialCombination(job.CharSet, job.KeyLen)
// 	i := 0
// 	for {
// 		newComb, rem := nextCombination(job.CharSet, size, comb)
// 		if rem != 0 {
// 			return
// 		}
// 		newTask := lib.Task{job.Job, i, newComb, size}
// 		tasks = append(tasks, newTask)
// 		i++
// 	}
// 	return tasks
// }

func nextCombination(charset lib.CharSet, v int, inp []byte) ([]byte, int) {
	base := len(charSetSlice[charset])
	res, rem := addToIntSlice(base, v, bytesToIntSlice(charset, inp))
	return intSliceToBytes(charset, res), rem
}

// TODO is combination the right term?
func initialCombination(charset lib.CharSet, keyLen int) []byte {
	v := charSetSlice[charset][0]
	res := make([]byte, keyLen)
	for i := range res {
		res[i] = v
	}
	return res
}

func bytesToIntSlice(charset lib.CharSet, inp []byte) []int {
	res := make([]int, len(inp))
	for i, b := range inp {
		x := bytes.IndexByte(charSetSlice[charset], b) // probably not efficient
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
	return inp, v
}

func intSliceToBytes(charset lib.CharSet, inp []int) []byte {
	res := make([]byte, len(inp))
	for i, v := range inp {
		res[i] = charSetSlice[charset][v]
	}
	return res
}

package master

import "github.com/aaronang/cong-the-ripper/lib"

// import "math.big"
import "bytes"

// must match lib.CharSet

type job struct {
	lib.Job
	id int
	// Tasks []lib.Task
}

// func SplitJob(job *job) []lib.Task {
// 	job.CharSet
// 	job.KeyLen
//
// }

func bytesToIntArray(charset lib.CharSet, inp []byte) []int {
	nums := []byte("0123456789")
	alphaLower := []byte("abcdefghijklmnopqrstuvwxyz")
	alphaMixed := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	alphaNumLower := []byte("0123456789abcdefghijklmnopqrstuvwxyz")
	charSetArray := [][]byte{nums, alphaLower, alphaMixed, alphaNumLower}

	res := make([]int, len(inp))
	for i, b := range inp {
		x := bytes.IndexByte(charSetArray[charset], b) // probably not efficient
		if x < 0 {
			panic("Invalid characters!")
		}
		res[i] = x
	}
	return res
}

func addToIntArray(base, v int, inp []int) []int {
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

	if v != 0 {
		panic("addToIntArray failed")
	}
	return inp
}

func intArrayToBytes(charset lib.CharSet, inp []int) []byte {
	// res := make([]byte, len(inp))
	return nil
}

// func strToBase10(charset lib.CharSet, inp string) big.Int {
//
// }

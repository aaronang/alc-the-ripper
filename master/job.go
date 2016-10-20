package master

import "github.com/aaronang/cong-the-ripper/lib"
import "fmt"

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
		x := bytes.IndexByte(charSetArray[charset], b)
		if x < 0 {
			panic("Invalid characters!")
		}
		res[i] = x
	}
	fmt.Println(res)
	return res
}

// func strToBase10(charset lib.CharSet, inp string) big.Int {
//
// }

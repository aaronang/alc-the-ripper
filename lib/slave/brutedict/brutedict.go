package brutedict

import (
	"log"
	"math/big"

	"github.com/aaronang/cong-the-ripper/lib"
)

type BruteDict struct {
	*lib.Task
	remaining   int
	currentComb *big.Int
}

func New(task *lib.Task) (bd *BruteDict) {
	bd = &BruteDict{
		Task:        task,
		remaining:   task.TaskLen,
		currentComb: lib.BytesToBigInt(task.Alphabet, task.Start),
	}

	if len(task.Progress) > 0 {
		progressComb := lib.BytesToBigInt(task.Alphabet, task.Progress)
		progressLen := big.NewInt(0)
		progressLen.Sub(progressComb, bd.currentComb)

		bd.remaining = bd.remaining - int(progressLen.Int64())
		if bd.remaining < 0 {
			log.Panic("[brutedict] bd.remaining is negative!", bd.remaining)
		}
		bd.currentComb = progressComb
	}

	return
}

func (bd *BruteDict) Next() (candidate []byte) {
	if bd.remaining <= 0 {
		return nil
	}

	pwLen := len(bd.Start)
	byteArray, overflow := lib.BigIntToBytes(bd.Alphabet, bd.currentComb, pwLen)
	if overflow {
		return nil
	}

	res := lib.ReverseArray(byteArray)
	bd.currentComb.Add(bd.currentComb, big.NewInt(1))
	bd.remaining--
	return res
}

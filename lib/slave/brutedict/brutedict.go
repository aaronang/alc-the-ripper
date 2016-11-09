package brutedict

import (
	"math/big"

	"github.com/aaronang/cong-the-ripper/lib"
)

type BruteDict struct {
	task  *lib.Task
	queue chan []byte
	done  chan bool
	kill  chan int
}

func New(task *lib.Task, killChan chan int) (bd *BruteDict) {
	bd = &BruteDict{
		task:  task,
		queue: make(chan []byte),
		done:  make(chan bool),
		kill:  killChan,
	}

	go bd.list(task.Alphabet, task.Start, task.Progress, task.TaskLen)
	return
}

func (bd *BruteDict) Next() (candidate []byte) {
	select {
	case candidate = <-bd.queue:
	case <-bd.done:
		close(bd.queue)
	}
	return
}

func (bd *BruteDict) list(alph lib.Alphabet, start []byte, progress []byte, length int) {
	currentComb := lib.BytesToBigInt(alph, start)
	if progress != nil {
		progressComb := lib.BytesToBigInt(alph, progress)
		progressLen := big.NewInt(0)
		progressLen.Sub(progressComb, currentComb)

		length = length - int(progressLen.Int64())
		currentComb = progressComb
	}

	pwLen := len(start)

outer:
	for length > 0 {
		select {
		case <-bd.kill:
			break outer
		default:
			byteArray, overflow := lib.BigIntToBytes(alph, currentComb, pwLen)
			if overflow {
				break outer
			}

			bd.queue <- lib.ReverseArray(byteArray)
			currentComb.Add(currentComb, big.NewInt(1))
			length--
		}
	}
	bd.done <- true
}

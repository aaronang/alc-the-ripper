package brutedict

import (
	"math/big"
	"sync"

	"github.com/aaronang/cong-the-ripper/lib"
)

type BruteDict struct {
	task    *lib.Task
	queue   chan []byte
	quit    chan bool
	stopped bool
	mutex   sync.RWMutex
}

func New(task *lib.Task) (bd *BruteDict) {
	bd = &BruteDict{
		task:    task,
		queue:   make(chan []byte),
		quit:    make(chan bool),
		stopped: false,
	}

	go bd.list(task.Alphabet, task.Start, task.Progress, task.TaskLen)
	return
}

func (bd *BruteDict) Next() (candidate []byte) {
	select {
	case candidate = <-bd.queue:
	case <-bd.quit:
	}
	return
}

func (bd *BruteDict) Close() {
	bd.mutex.Lock()
	bd.stopped = true
	bd.mutex.Unlock()
	close(bd.queue)
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

	len := len(start)

	for length > 0 {
		bd.mutex.RLock()
		if bd.stopped {
			bd.mutex.RUnlock()
			break
		}
		bd.mutex.RUnlock()

		byteArray, overflow := lib.BigIntToBytes(alph, currentComb, len)
		if overflow {
			break
		}

		bd.queue <- lib.ReverseArray(byteArray)
		currentComb.Add(currentComb, big.NewInt(1))
		length--
	}

	bd.quit <- true
}

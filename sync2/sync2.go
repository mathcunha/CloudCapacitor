package sync2

import (
	"errors"
	"sync"
)

type BlockWaitGroup struct {
	max     int
	current int
	wg      *sync.WaitGroup
}

func NewBlockWaitGroup(size int) (bwg *BlockWaitGroup) {
	bwg = new(BlockWaitGroup)
	bwg.max = size
	bwg.wg = new(sync.WaitGroup)
	return bwg
}

func (bwg *BlockWaitGroup) Add(delta int) (current int, err error) {
	if bwg.current <= bwg.max {
		bwg.current++
		bwg.wg.Add(delta)
	} else {
		return bwg.current, errors.New("max goroutines reached")
	}
	return bwg.current, nil
}

func (bwg *BlockWaitGroup) Done() {
	bwg.current--
	bwg.wg.Done()
}

func (bwg *BlockWaitGroup) Wait() {
	bwg.wg.Wait()
}

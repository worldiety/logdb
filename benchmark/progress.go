package benchmark

import (
	"fmt"
	"sync/atomic"
	"time"
)

// Progress estimates duration and prints some statistics
type Progress struct {
	max                uint64
	current            uint64
	lastTick           int64
	startedAt          int64
	checkElemsInterval uint64
}

// NewProgress starts a progress now
func NewProgress(max, checkStartInterval uint64) *Progress {
	return &Progress{
		max:                max,
		current:            0,
		startedAt:          time.Now().Unix(),
		checkElemsInterval: checkStartInterval,
	}
}

// Next increments the progress and is safe to be called concurrently
func (p *Progress) Next() {

	current := atomic.AddUint64(&p.current, 1)
	interval := atomic.LoadUint64(&p.checkElemsInterval)
	if current%interval == 0 {
		if time.Now().Sub(time.Unix(atomic.LoadInt64(&p.lastTick), 0)) < 30*time.Second {
			atomic.StoreUint64(&p.checkElemsInterval, interval*2) // this is logically racy, but not so important
			return
		}

		atomic.StoreInt64(&p.lastTick, time.Now().Unix())
		p.PrintStatistics()
	}
}

// PrintStatistics just prints some global statistics
func (p *Progress) PrintStatistics() {
	elements := atomic.LoadUint64(&p.current)
	startedAt := atomic.LoadInt64(&p.startedAt)
	duration := time.Now().Sub(time.Unix(startedAt, 0))
	durationPerElem := float64(duration.Nanoseconds()) / float64(elements)
	elemsPerSecond := float64(elements) / duration.Seconds()

	fmt.Printf("-------\n")
	fmt.Printf("%2.f%%\n", float64(elements)/float64(p.max)*100)
	fmt.Printf("processed %d elements in %v\n", elements, duration)
	fmt.Printf("processing duration per element was %v\n", time.Duration(durationPerElem))
	fmt.Printf("processed %.2f elements per second", elemsPerSecond)
	if elemsPerSecond > 100_000 {
		fmt.Printf(" (%.f million elements per second)", elemsPerSecond/1_000_000)
	}
	fmt.Printf("\n")
	fmt.Printf("\n")
}

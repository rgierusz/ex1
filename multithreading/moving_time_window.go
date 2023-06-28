package multithreading

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type valueWithTs struct {
	v    int
	tsMs int64
}

func (v valueWithTs) String() string {
	return fmt.Sprintf("%v", v.v)
}

type MovingWindow struct {
	periodMs   int64
	values     []valueWithTs
	mu         sync.RWMutex
	processors []func([]valueWithTs) float64
}

func NewMovingWindow(periodMs int64, processors ...func([]valueWithTs) float64) *MovingWindow {
	return &MovingWindow{
		periodMs:   periodMs,
		values:     make([]valueWithTs, 0),
		processors: processors,
	}
}

func (mw *MovingWindow) AddValue(value int) {
	mw.mu.Lock()
	defer mw.mu.Unlock()

	mw.removeOutdatedValues()

	mw.values = append(mw.values, valueWithTs{
		v:    value,
		tsMs: time.Now().UnixMilli(),
	})
}

func (mw *MovingWindow) ProcessWithProcessors() []float64 {
	mw.mu.RLock()
	defer mw.mu.RUnlock()

	values, _ := mw.getValidValues()

	var results []float64
	for _, processor := range mw.processors {
		results = append(results, processor(values))
	}

	return results
}

// this function assumes that the locking is external
// the bool part of the result indicates if the returned value is different from the internal window's value
func (mw *MovingWindow) getValidValues() ([]valueWithTs, bool) {
	if len(mw.values) == 0 {
		return mw.values, false
	}

	edgeTs := time.Now().UnixMilli() - mw.periodMs

	lastIndexToSkip := -1
	for i, v := range mw.values {
		if v.tsMs >= edgeTs {
			break
		}

		lastIndexToSkip = i
	}

	if lastIndexToSkip == -1 {
		return mw.values, false // exactly the same slice, nothing to skip
	} else if lastIndexToSkip == len(mw.values)-1 {
		return mw.values[:0], true // everything was skipped, empty slice
	}

	return mw.values[lastIndexToSkip+1:], true
}

// this function assumes that the locking is external
func (mw *MovingWindow) removeOutdatedValues() {
	if values, changed := mw.getValidValues(); changed {
		mw.values = values
	}
}

// AverageProcessor example processor to demonstrate external usage
func AverageProcessor(values []valueWithTs) float64 {
	var sum int
	count := len(values)

	for _, value := range values {
		sum += value.v
	}

	if count > 0 {
		average := float64(sum) / float64(count)
		log.Printf("Processor 1: count %v -> %v -> average: %v\n", len(values), values, average)

		return average
	}

	log.Printf("Processor 1: count %v -> %v -> average: %v", len(values), values, 0.0)
	return 0.0
}

// SumProcessor example processor to demonstrate external usage
func SumProcessor(values []valueWithTs) float64 {
	var sum int
	for _, value := range values {
		sum += value.v
	}

	log.Printf("Processor 2: sum: %v\n", sum)
	return float64(sum)
}

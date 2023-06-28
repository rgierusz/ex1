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
	periodMs  int64
	values    []valueWithTs
	mu        sync.RWMutex
	processor func([]valueWithTs) float64
}

func NewMovingWindow(periodMs int64, processor func([]valueWithTs) float64) *MovingWindow {
	return &MovingWindow{
		periodMs:  periodMs,
		values:    make([]valueWithTs, 0),
		processor: processor,
	}
}

func (mw *MovingWindow) AddValue(value int) {
	mw.mu.Lock()
	defer mw.mu.Unlock()

	currentTsMs := time.Now().UnixMilli()

	mw.removeOutdatedValues(currentTsMs)

	mw.values = append(mw.values, valueWithTs{
		v:    value,
		tsMs: currentTsMs,
	})
}

func (mw *MovingWindow) CalculateAverage() float64 {
	mw.mu.RLock()
	defer mw.mu.RUnlock()

	return mw.processor(mw.getValidValuesSlice())
}

// there is no modification of values performed here
// can be combined with removeOutdatedValues method, but for the reason of clarity  - and performance - it's implemented separately
func (mw *MovingWindow) getValidValuesSlice() []valueWithTs {
	edgeTs := time.Now().UnixMilli() - mw.periodMs
	toSkipIndex := -1

	for i, v := range mw.values {
		if v.tsMs < edgeTs {
			toSkipIndex = i
		} else {
			break
		}
	}

	if toSkipIndex == -1 {
		return mw.values
	}

	return mw.values[toSkipIndex+1:]
}

// this function assumed that the locking is external
func (mw *MovingWindow) removeOutdatedValues(currentTsMs int64) {
	if len(mw.values) == 0 {
		return
	}

	edgeTs := currentTsMs - mw.periodMs

	for i, v := range mw.values {
		if v.tsMs >= edgeTs { // current value is not outdated, no need to continue iteration
			if i == 0 { // current (not outdated) value is the first one, no need to remove anything
				return
			}

			mw.values = mw.values[i:] // remove all values before the current value
			return
		}
	}

	mw.values = mw.values[:0] // all values are outdated and need to be removed
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
		log.Printf("Count %v -> %v -> average: %v", len(values), values, average)

		return average
	}

	log.Printf("Count %v -> %v -> average: %v", len(values), values, 0.0)
	return 0.0
}

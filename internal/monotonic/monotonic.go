package monotonic

import (
	"log"
	"time"
)

type Fingerprint struct {
	startTime     time.Time
	monotonicBase int64
	tickHistory   []time.Duration
	expectedTicks int64
}

func NewFingerprint() *Fingerprint {
	n := time.Now()
	return &Fingerprint{
		startTime:     n,
		monotonicBase: n.UnixNano(),
		tickHistory:   make([]time.Duration, 0, 1000),
		expectedTicks: 0,
	}
}

func (f *Fingerprint) Record() {
	n := time.Now()
	elapsed := n.Sub(f.startTime)
	f.tickHistory = append(f.tickHistory, elapsed)
	f.expectedTicks++
}

func (f *Fingerprint) Detect() bool {
	// low sample volume
	if len(f.tickHistory) < 100 {
		// less noisy logging
		if len(f.tickHistory)%10 == 0 {
			log.Println("monotonic - low sample count skipping check", len(f.tickHistory))
		}
		return false
	}

	// check for time going backwards - indicative of a snapshot restoration
	for i := 1; i < len(f.tickHistory); i++ {
		if f.tickHistory[i] < f.tickHistory[i-1] {
			return true
		}
	}

	// check for sudden time jumps
	recent := f.tickHistory[len(f.tickHistory)-10:]
	var avgDelta time.Duration
	for i := range len(recent) {
		if i != 0 {
			avgDelta += recent[i] - recent[i-1]
		}
	}

	avgDelta /= time.Duration(len(recent) - 1)
	log.Println("monotonic - average delta:", avgDelta)

	// comp latest time delta to recent average
	lastDelta := recent[len(recent)-1] - recent[len(recent)-2]
	log.Println("monotonic - last delta:", lastDelta)

	ratio := float64(lastDelta) / float64(avgDelta)
	log.Println("monotonic - ratio:", ratio)

	// sudden large jumps can be an indicator of a snapshot restoration
	return ratio > 10.0 || ratio < 0.1
}

package main

import (
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	"code.t25.tokyo/lockdown/internal/entropy"
	"code.t25.tokyo/lockdown/internal/layout"
	"code.t25.tokyo/lockdown/internal/monotonic"
	"code.t25.tokyo/lockdown/internal/timing"
)

func main() {
	log.Println("setting up monitoring")

	go NewMonitor()

	runtime.Goexit()
}

type Monitor struct {
	timingFP    *timing.Fingerprint
	layoutFP    *layout.Fingerprint
	monotonicFP *monotonic.Fingerprint
	entropyFP   *entropy.Fingerprint
	threshold   int
	mu          sync.RWMutex
}

func NewMonitor() *Monitor {
	m := &Monitor{
		entropyFP:   entropy.NewFingerprint(),
		layoutFP:    layout.NewFingerprint(),
		monotonicFP: monotonic.NewFingerprint(),
		timingFP:    timing.NewFingerprint(),
		threshold:   3,
	}

	go m.poll()

	return m
}

func (m *Monitor) poll() {
	tick := time.NewTicker(time.Second * 5)
	defer tick.Stop()

	log.Println("starting monitoring")

	for range tick.C {
		m.monotonicFP.Record()
		if m.assert() {
			log.Println("critical - potential cloning/snapshot detected")
			// Handle graceful actions
			os.Exit(1)
		}
	}
}

func (m *Monitor) assert() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	detections := 0

	if m.entropyFP.Detect() {
		detections++
		log.Println("inconsistency detected in pRNG")
	}

	if m.layoutFP.Detect() {
		detections++
		log.Println("inconsistency detected in memory allocation")
	}

	if m.monotonicFP.Detect() {
		detections++
		log.Println("inconsistency detected in system tick")
	}

	if m.timingFP.Detect() {
		detections++
		log.Println("inconsistency detected in cpu timing")
	}

	log.Printf("indicators found: %d", detections)
	return detections >= m.threshold
}

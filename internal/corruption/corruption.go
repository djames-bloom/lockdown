package corruption

import (
	"log"
	"runtime"
	"sync"
	"time"
)

type Fingerprint struct {
	mu           sync.Mutex
	allocs       int64
	heapObjects  uint64
	lastSnapshot time.Time
	stopChan     chan struct{}
}

func NewFingerprint() *Fingerprint {
	fp := &Fingerprint{
		lastSnapshot: time.Now(),
		stopChan:     make(chan struct{}),
	}

	// initial snapshot
	fp.Record()
	fp.Corrupt()

	return fp
}

func (f *Fingerprint) Record() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.updateStats()
}

func (f *Fingerprint) updateStats() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	f.heapObjects = ms.HeapObjects
	f.allocs = int64(ms.Mallocs)
	f.lastSnapshot = time.Now()
}

func (f *Fingerprint) Detect() bool {
	if !f.mu.TryLock() {
		log.Println("corruption - lock not released, skipping this cycle")
		return false
	}
	defer f.mu.Unlock()

	prevAllocs := f.allocs
	prevHeapObjects := f.heapObjects

	// Update stats directly instead of calling Record()
	f.updateStats()

	// resources should generally increase over time as the application
	// executes, if they decrease significantly it may indicate a rollback
	// has been applied
	allocDelta := f.allocs - prevAllocs
	heapDelta := f.heapObjects - prevHeapObjects

	log.Println("corruption - alloc delta: ", allocDelta)
	log.Println("corruption - heap delta: ", heapDelta)

	// suspicious delta in allocations
	if allocDelta < -1000 && f.allocs > 0 {
		log.Println("corruption - inconsistency detected in allocations")
		return true
	}

	// suspicious large reduction in heap objects
	if f.heapObjects < prevHeapObjects/2 &&
		prevHeapObjects > 1000 {
		log.Println("corruption - inconsistency detected in heap size")
		return true
	}

	return false
}

// Corrupt creates goroutines that gradually soak resources that should be
// lost if a snapshot/rollback is applied
func (f *Fingerprint) Corrupt() {
	for range 10 {
		go func() {
			tick := time.NewTicker(time.Second)
			defer tick.Stop()

			var buf []byte
			for {
				select {
				case <-tick.C:
					// buffer for gradual soak
					buf = append(buf, make([]byte, 1024)...)

					// cap at 1mb consumed per goroutine
					if len(buf) > 1024*1024 {
						buf = buf[:0]
					}
				case <-f.stopChan:
					return
				}
			}
		}()
	}
}

func (f *Fingerprint) Stop() {
	close(f.stopChan)
}

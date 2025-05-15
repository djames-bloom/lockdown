package timing

import (
	"log"
	"runtime"
	"time"
)

type Fingerprint struct {
	baselineLatency time.Duration
	cpuFeatures     string
}

const (
	iterations = 100000
)

func NewFingerprint() *Fingerprint {
	start := time.Now()
	for i := range iterations {
		_ = i
		runtime.Gosched() // force context switches
	}
	baseline := time.Since(start)

	return &Fingerprint{
		baselineLatency: baseline,
		cpuFeatures:     runtime.GOARCH + runtime.GOOS,
	}
}

// Detect repeats the initial timing fingerprinting process and compares
// the scheduling delta. Cloned VMs often have poor clock timing consistency
// and can be used as an indicator of tampering. It is however not reliable
// as a detection alone and should be used as a 1 of n indicator
func (f *Fingerprint) Detect() bool {
	// simplest check first - assert matching cpu features
	if f.cpuFeatures != runtime.GOARCH+runtime.GOOS {
		log.Println("timing - cpu features mismatch: ", f.cpuFeatures, runtime.GOARCH+runtime.GOOS)
		return true
	}

	start := time.Now()
	for i := range iterations {
		_ = i
		runtime.Gosched()
	}
	current := time.Since(start)

	ratio := float64(current) / float64(f.baselineLatency)

	log.Println("timing - ratio: ", ratio)

	// Imaged VMs often have timing inconsistencies
	// If the delta is significant the VM may have been cloned.
	return ratio < 0.5 || ratio > 2.0
}

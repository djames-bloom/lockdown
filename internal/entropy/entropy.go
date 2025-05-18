package entropy

import (
	"crypto/rand"
	"log"
	"math/big"
	"time"
)

type Fingerprint struct {
	seedHistory []int64
	genRate     time.Duration
	entropyPool []byte
}

func NewFingerprint() *Fingerprint {
	fp := &Fingerprint{
		seedHistory: make([]int64, 0, 100),
		entropyPool: make([]byte, 1024),
	}

	// initial snapshot
	rand.Read(fp.entropyPool)

	return fp
}

func (f *Fingerprint) Record() {
	start := time.Now()

	for i := 0; i < 10; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
		f.seedHistory = append(f.seedHistory, n.Int64())
	}

	f.genRate = time.Since(start)
}

func (f *Fingerprint) Detect() bool {
	previousRate := f.genRate
	f.Record()

	// low sample volume
	if len(f.seedHistory) < 20 {
		log.Println("entropy - low seed pool")
		return false
	}

	// observe repetititive patterns - indicative of a snapshot restoration
	recent := f.seedHistory[len(f.seedHistory)-10:]
	for i := 0; i < len(f.seedHistory)-10; i++ {
		window := f.seedHistory[i : i+10]
		if f.sequencesMatch(recent, window) {
			log.Println("entropy - found a matching value")
			// repetitions found
			return true
		}
	}

	ratio := float64(f.genRate) / float64(previousRate)
	log.Printf("entropy - ratio: %v", ratio)

	return ratio < 0.1 || ratio > 10.0
}

func (f *Fingerprint) sequencesMatch(a, b []int64) bool {
	if len(a) != len(b) {
		return false
	}
	matches := 0
	for i := range a {
		if a[i] == b[i] {
			matches++
		}
	}

	// round up 75% match rate
	return matches > 8
}

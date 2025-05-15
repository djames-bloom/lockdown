package layout

import (
	"bytes"
	"log"
	"runtime"
	"strconv"
	"unsafe"
)

type Fingerprint struct {
	stackBase    uintptr
	heapBase     uintptr
	goroutineID  uint64
	memoryLayout [10]uintptr
}

func NewFingerprint() *Fingerprint {
	var buf [10]uintptr

	for i := range buf {
		var local int
		buf[i] = uintptr(unsafe.Pointer(&local))
	}

	return &Fingerprint{
		stackBase:    uintptr(unsafe.Pointer(&buf)),
		heapBase:     getHeapBase(),
		goroutineID:  getGoroutineID(),
		memoryLayout: buf,
	}
}

func (f *Fingerprint) Detect() bool {
	current := NewFingerprint()

	// memory layout should differ between executions
	// if they're identical its a good sign we've been
	// cloned
	matches := 0
	for i := range f.memoryLayout {
		if f.memoryLayout[i] == current.memoryLayout[i] {
			matches++
		}
	}

	// log.Println("baseline:", f.memoryLayout)
	// log.Println("current:", current.memoryLayout)
	log.Println("layout - matches: ", matches)

	// set threshold to reduce false positives
	return matches > 7
}

func getGoroutineID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	// Stack trace format: "goroutine 123 [running]:"
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

func getHeapBase() uintptr {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return uintptr(ms.HeapSys)
}

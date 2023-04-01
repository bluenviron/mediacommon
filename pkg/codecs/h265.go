package codecs

import (
	"sync"
)

// H265 is a H265 codec.
type H265 struct {
	VPS []byte
	SPS []byte
	PPS []byte

	mutex sync.RWMutex
}

// SafeVPS returns the VPS.
func (t *H265) SafeVPS() []byte {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.VPS
}

// SafeSPS returns the SPS.
func (t *H265) SafeSPS() []byte {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.SPS
}

// SafePPS returns the PPS.
func (t *H265) SafePPS() []byte {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.PPS
}

// SafeSetVPS sets the VPS.
func (t *H265) SafeSetVPS(v []byte) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.VPS = v
}

// SafeSetSPS sets the SPS.
func (t *H265) SafeSetSPS(v []byte) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.SPS = v
}

// SafeSetPPS sets the PPS.
func (t *H265) SafeSetPPS(v []byte) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.PPS = v
}

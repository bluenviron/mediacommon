package codecs

import (
	"sync"
)

// H264 is a H264 codec.
type H264 struct {
	SPS []byte
	PPS []byte

	mutex sync.RWMutex
}

// SafeSPS returns the SPS.
func (t *H264) SafeSPS() []byte {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.SPS
}

// SafePPS returns the PPS.
func (t *H264) SafePPS() []byte {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.PPS
}

// SafeSetSPS sets the SPS.
func (t *H264) SafeSetSPS(v []byte) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.SPS = v
}

// SafeSetPPS sets the PPS.
func (t *H264) SafeSetPPS(v []byte) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.PPS = v
}

package providers

import (
	"math/rand"
	"sync"
)

type SessionIDProvider struct {
	initialOffset uint32
	mutex         sync.Mutex
}

func NewSessionIDProvider() *SessionIDProvider {
	return &SessionIDProvider{initialOffset: rand.Uint32()}
}

func (p *SessionIDProvider) GetNextSessionID() uint32 {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.initialOffset += 1
	return p.initialOffset
}

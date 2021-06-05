package providers

import (
	"math/rand"
	"sync"
)

type SessionIDProvider struct {
	initialOffset int32
	mutex         sync.Mutex
}

func NewSessionIDProvider() *SessionIDProvider {
	return &SessionIDProvider{initialOffset: rand.Int31()}
}

func (p *SessionIDProvider) GetNextSessionID() int32 {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.initialOffset += 1
	return p.initialOffset
}

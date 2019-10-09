package mem

import (
	"context"
	"sync"
)

type Memory struct {
	rw sync.RWMutex

	content []byte
	ver     int64
	c       chan struct{}
}

func (m *Memory) Set(newc []byte) {
	m.rw.Lock()
	close(m.c)
	m.c = make(chan struct{})
	m.content = newc
	m.ver++
	m.rw.Unlock()
}

func (m *Memory) Next(ctx context.Context, oldversion int64) (content []byte, curversion int64, ok bool) {
	m.rw.RLock()
	if m.ver != oldversion {
		defer m.rw.RUnlock()
		content, curversion, ok = m.content, m.ver, true
		return
	}
	copyc := m.c
	m.rw.RUnlock()
	select {
	case <-ctx.Done():
		return
	case <-copyc:
		m.rw.RLock()
		content, curversion, ok = m.content, m.ver, true
		m.rw.RUnlock()
		return
	}
}

func (m *Memory) Close() {
	m.rw.Lock()
	defer m.rw.Unlock()
	select {
	case <-m.c:
		return
	default:
		close(m.c)
	}
}

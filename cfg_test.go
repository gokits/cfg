package cfg

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

type MockSource struct {
	ver     int64
	content []byte
}

func (ms *MockSource) Next(ctx context.Context, oldversion int64) (content []byte, version int64, ok bool) {
	if oldversion != ms.ver {
		content, version, ok = ms.content, ms.ver, true
		return
	}
	select {
	case <-ctx.Done():
		return
	}
}

func (ms *MockSource) Close() {

}

type Config struct {
	A string
}

var PostDecodeCnt int64

func (cc *Config) PostDecode(c interface{}) error {
	atomic.AddInt64(&PostDecodeCnt, 1)
	time.Sleep(time.Second)
	return errors.New("dddd")
}

func TestDeadLoop(t *testing.T) {
	s := &MockSource{
		content: []byte(`{"A":"ddd"}`),
		ver:     1,
	}
	meta := NewConfigMeta(Config{}, s)
	go meta.Run()
	defer meta.Stop()
	time.Sleep(15 * time.Second)
	loopcnt := atomic.LoadInt64(&PostDecodeCnt)
	if loopcnt != 1 {
		t.Fatalf("Deadloop detected, loopcnt=%d\n", loopcnt)
	}
}

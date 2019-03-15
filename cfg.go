package cfg

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"time"
)

type Config interface {
	PreLoad(cfgptr interface{}) error
	PostLoad(cfgptr interface{}) error
}

type ConfigMeta struct {
	ct       reflect.Type
	rw       sync.RWMutex
	instance interface{}
	synced   bool
	version  int64
	source   Source
	decoder  Decoder
	stopped  chan int
}

type Option func(cm *ConfigMeta)

func WithDecoder(d Decoder) Option {
	return func(cm *ConfigMeta) {
		cm.decoder = d
	}
}

func NewConfigMeta(c interface{}, source Source, opts ...Option) *ConfigMeta {
	cm := &ConfigMeta{
		ct:      reflect.TypeOf(c),
		decoder: new(JsonDecoder),
		source:  source,
	}
	if cm.ct.Kind() == reflect.Ptr {
		cm.ct = cm.ct.Elem()
	}
	for _, opt := range opts {
		opt(cm)
	}
	return cm
}

func (cm *ConfigMeta) Run() {
	var err error
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		data, curversion, ok := cm.source.Next(ctx, cm.version)
		cancel()
		select {
		case <-cm.stopped:
			return
		default:
			if ok {
				ncv := reflect.New(cm.ct)
				nc := ncv.Interface().(Config)
				if err = nc.PreLoad(ncv.Interface()); err != nil {
					//TODO log this
					continue
				}
				if err = cm.decoder.Unmarshal(data, ncv.Interface()); err != nil {
					//TODO log this
					continue
				}
				if err = nc.PostLoad(ncv.Interface()); err != nil {
					//TODO log this
					continue
				}
				cm.rw.Lock()
				cm.instance = ncv.Interface()
				cm.version = curversion
				cm.synced = true
				cm.rw.Unlock()
			}
		}
	}
}

func (cm *ConfigMeta) WaitSynced() error {
	for {
		select {
		case <-cm.stopped:
			return errors.New("stopped")
		default:
			if cm.synced {
				return nil
			}
			time.Sleep(time.Second)
			continue
		}
	}
}

func (cm *ConfigMeta) Get() interface{} {
	cm.rw.RLock()
	defer cm.rw.RUnlock()
	return cm.instance
}

func (cm *ConfigMeta) Stop() {
	select {
	case <-cm.stopped:
		return
	default:
		close(cm.stopped)
	}
}

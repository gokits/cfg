package cfg

import (
	"sync"
	"os"
	"encoding/json"
	"reflect"
	"sync/atomic"
    "github.com/fsnotify/fsnotify"
)

type Config interface {
	PreLoad(cfgptr interface{}) error
	PostLoad(cfgptr interface{}) error
}

type Source interface {
	Watch() chan []byte
	Close()
}

type File struct {
    filename string
    watcher *fsnotify.Watcher
    contentrw sync.RWMutex
    content []byte
    c chan struct{}
}

func NewFileSource(filename string) (fs *File, err error) {
    fs = &{
        filename: filename,
    }
    if fs.watcher, err = fsnotify.NewWatcher(); err != nil {
        return
    }
    defer func() {
        if err != nil {
            fs.watcher.Close()
        }
    }()
    if err = fs.watcher.Add(filename); err != nil {
        return
    }
    fs.c = make(chan struct{})
    return
}

func (f *File) WatchSignal() chan struct{}  {
    return f.c
}

func (f *File) Close() {
    f.watcher.Close()
    close(f.c)
}


type Decoder interface {
	Unmarshal(data []byte, v interface{}) error
}

type JsonDecoder int

func (jd *JsonDecoder) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

type ConfigMeta struct {
	ct       reflect.Type
	instance atomic.Value
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

func (cm *ConfigMeta) Start() {
	var err error
	next := cm.source.Watch()
	for {
		select {
		case data, ok := <-next:
			if !ok {
				return
			}
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
			cm.instance.Store(ncv.Interface())
		case <-cm.stopped:
			return
		}
	}
}

func (cm *ConfigMeta) Get() interface{} {
	return cm.instance.Load()
}

func (cm *ConfigMeta) Stop() {
	select {
	case <-cm.stopped:
		return
	default:
		close(cm.stopped)
	}
}

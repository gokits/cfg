package cfg

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/gokits/cfg/decoder/json"
	logger "github.com/gokits/stdlogger"
	nooplogger "github.com/gokits/stdlogger/nooplogger"
)

type PreDecoder interface {
	PreDecode(oldptr interface{}) error
}

type PostDecoder interface {
	PostDecode(oldptr interface{}) error
}

type PostSwapper interface {
	PostSwap(oldptr interface{})
}

type ConfigMeta struct {
	ct            reflect.Type
	rw            sync.RWMutex
	instance      interface{}
	logger        logger.LeveledLogger
	validatorInst *validator.Validate
	synced        bool
	version       int64
	source        Source
	decoder       Decoder
	stopped       chan int
}

type Option func(cm *ConfigMeta)

func WithDecoder(d Decoder) Option {
	return func(cm *ConfigMeta) {
		cm.decoder = d
	}
}

func WithLogger(logger logger.LeveledLogger) Option {
	return func(cm *ConfigMeta) {
		cm.logger = logger
	}
}

func WithValidator(v *validator.Validate) Option {
	return func(cm *ConfigMeta) {
		cm.validatorInst = v
	}
}

func NewConfigMeta(c interface{}, source Source, opts ...Option) *ConfigMeta {
	cm := &ConfigMeta{
		ct:            reflect.TypeOf(c),
		decoder:       new(json.JsonDecoder),
		source:        source,
		stopped:       make(chan int),
		logger:        nooplogger.Default(),
		validatorInst: validator.New(),
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
	var (
		ok          bool
		err         error
		tmpv        int64
		predecoder  PreDecoder
		postdecoder PostDecoder
		postswapper PostSwapper
		old         interface{}
	)
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		var data []byte
		data, tmpv, ok = cm.source.Next(ctx, cm.version)
		cancel()
		select {
		case <-cm.stopped:
			return
		default:
			if ok {
				cm.version = tmpv
				ncv := reflect.New(cm.ct)
				if predecoder, ok = ncv.Interface().(PreDecoder); ok {
					if err = predecoder.PreDecode(cm.instance); err != nil {
						cm.logger.Infof("PreDecode error: %v", err)
						continue
					}
				}

				if err = cm.decoder.Unmarshal(data, ncv.Interface()); err != nil {
					cm.logger.Warnf("Unmarshal error: %v, data: %s", err, string(data))
					continue
				}
				if cm.validatorInst != nil {
					if err = cm.validatorInst.Struct(ncv.Interface()); err != nil {
						cm.logger.Warnf("Values invalid with validator: %v", err)
						continue
					}
				}
				if postdecoder, ok = ncv.Interface().(PostDecoder); ok {
					if err = postdecoder.PostDecode(cm.instance); err != nil {
						cm.logger.Infof("PostDecode error: %v", err)
						continue
					}
				}
				cm.rw.Lock()
				old = cm.instance
				cm.instance = ncv.Interface()
				cm.synced = true
				cm.rw.Unlock()
				cm.logger.Infof("success swap config. version: %d", cm.version)
				if postswapper, ok = ncv.Interface().(PostSwapper); ok {
					postswapper.PostSwap(old)
				}
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
			cm.rw.RLock()
			if cm.synced {
				cm.rw.RUnlock()
				return nil
			}
			cm.rw.RUnlock()
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

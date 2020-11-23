package cfg

import (
	"fmt"
	"reflect"
	"sync"

	validator "github.com/go-playground/validator/v10"
	"github.com/gokits/cfg/decoder/json"
	filesource "github.com/gokits/cfg/source/file"
	logger "github.com/gokits/stdlogger"
	nooplogger "github.com/gokits/stdlogger/nooplogger"
)

var (
	instances *sync.Map = &sync.Map{}
)

type instanceValue struct {
	meta *ConfigMeta
	src  Source
}

func Final() {
	instances.Range(func(key, value interface{}) bool {
		value.(*instanceValue).src.Close()
		value.(*instanceValue).meta.Stop()
		return true
	})
}

type RegisterConfiguration struct {
	decoder   Decoder
	logger    logger.LeveledLogger
	validator *validator.Validate
}

func (c *RegisterConfiguration) WithDecoder(d Decoder) *RegisterConfiguration {
	c.decoder = d
	return c
}

func (c *RegisterConfiguration) WithLogger(l logger.LeveledLogger) *RegisterConfiguration {
	c.logger = l
	return c
}

func (c *RegisterConfiguration) WithValidator(v *validator.Validate) *RegisterConfiguration {
	c.validator = v
	return c
}

func WithDefaultConfiguration() *RegisterConfiguration {
	return &RegisterConfiguration{
		decoder:   new(json.JsonDecoder),
		logger:    nooplogger.Default(),
		validator: validator.New(),
	}
}

func MustRegisterFile(configptr interface{}, file string, conf *RegisterConfiguration) {
	var err error
	v := &instanceValue{}
	v.src, err = filesource.NewFileSource(file, filesource.WithLogger(conf.logger))
	if err != nil {
		conf.logger.Errorf("new filesource %s failed: %v", file, err)
		panic(err)
	}
	v.meta = NewConfigMeta(configptr, v.src, WithDecoder(conf.decoder),
		WithLogger(conf.logger),
		WithValidator(conf.validator),
	)
	_, loaded := instances.LoadOrStore(reflect.TypeOf(configptr), v)
	if loaded {
		conf.logger.Errorf("config with type %v already registered", reflect.TypeOf(configptr))
		panic(fmt.Errorf("config with type %v already registered", reflect.TypeOf(configptr)))
	}
	go v.meta.Run()
	return
}

func MustGet(configptr interface{}) interface{} {
	key := reflect.TypeOf(configptr)
	v, ok := instances.Load(key)
	if !ok {
		panic(fmt.Errorf("config type %v not registered", key))
	}
	return v.(*instanceValue).meta.Get()
}

func WaitSynced(configptr interface{}) error {
	key := reflect.TypeOf(configptr)
	v, ok := instances.Load(key)
	if !ok {
		return fmt.Errorf("config type %v not registered", key)
	}
	return v.(*instanceValue).meta.WaitSynced()
}

func WaitSyncedAll() (err error) {
	instances.Range(func(k, v interface{}) bool {
		if err = v.(*instanceValue).meta.WaitSynced(); err != nil {
			return false
		}
		return true
	})
	return
}

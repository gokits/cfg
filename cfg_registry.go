package cfg

import (
	"fmt"
	"sync"
)

var (
	configRegistry sync.Map
)

func MustRegister(name string, meta *ConfigMeta) {
	c, exist := configRegistry.LoadOrStore(name, meta)
	if exist {
		panic(fmt.Errorf("config name %s already registered", name))
	}
	if err := meta.source.Start(); err != nil {
		panic(fmt.Errorf("source %s start failed: %v", meta.source, err))
	}
	go c.(*ConfigMeta).Run()
}

func WaitSyncedAll() (err error) {
	configRegistry.Range(func(k, v interface{}) bool {
		if err = v.(*ConfigMeta).WaitSynced(); err != nil {
			return false
		}
		return true
	})
	return
}

func MustGet(name string) interface{} {
	return MustGetConfigMeta(name).Get()
}

func MustGetConfigMeta(name string) *ConfigMeta {
	v, ok := configRegistry.Load(name)
	if !ok {
		panic(fmt.Errorf("config name %s not registered", name))
	}
	return v.(*ConfigMeta)
}

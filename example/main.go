package main

import (
	"errors"
	"flag"
	"fmt"
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/gokits/cfg"
	"github.com/gokits/cfg/source/file"
	stdlogrus "github.com/gokits/stdlogger/logrus"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Mode    string `validate:"oneof=file net"`
	SnapLen int    `validate:"min=32,max=1024000"`
}

var (
	filesource *file.File
	cfgfile    string
	meta       *cfg.ConfigMeta
	gvalidator *validator.Validate
)

func (c *Config) PostDecode(oldptr interface{}) error {
	var err error
	if err = gvalidator.Struct(c); err != nil {
		return errors.New(fmt.Sprintf("validate config failed: %+v", err))
	}
	return nil
}

func (c *Config) PostSwap(oldptr interface{}) {
	logrus.Infof("config swapped: %+v", c)
}

func main() {
	flag.StringVar(&cfgfile, "cfg", "./test.json", "file path of config file")
	flag.Parse()

	var err error

	if filesource, err = file.NewFileSource(cfgfile, file.WithLogger(stdlogrus.FromGlobal())); err != nil {
		logrus.Fatal(err)
	}
	meta = cfg.NewConfigMeta(Config{}, filesource, cfg.WithLogger(stdlogrus.FromGlobal()))
	gvalidator = validator.New()
	defer Fini()
	go meta.Run()
	if err = meta.WaitSynced(); err != nil {
		filesource.Close()
		logrus.WithError(err).Errorf("waitsync failed")
		return
	}
	logrus.Infof("after waitsynced: %+v", Get())
	go func() {
		for {
			time.Sleep(3 * time.Second)
			logrus.Info(Get())
		}
	}()
	time.Sleep(time.Hour)
	return
}

func Fini() {
	meta.Stop()
	filesource.Close()
}

func Get() *Config {
	return meta.Get().(*Config)
}

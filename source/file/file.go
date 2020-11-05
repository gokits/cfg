package file

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gokits/cfg"
	logger "github.com/gokits/stdlogger"
	"github.com/gokits/stdlogger/nooplogger"
)

type File struct {
	filename string
	watcher  *fsnotify.Watcher
	rw       sync.RWMutex
	logger   logger.LeveledLogger
	content  []byte
	ver      int64
	c        chan struct{}
}

type Option func(*File)

func WithLogger(log logger.LeveledLogger) Option {
	return func(f *File) {
		f.logger = log
	}
}

func WithFilePath(filepath string) Option {
	return func(f *File) {
		f.filename = filepath
	}
}

func NewFileSource(filepath string, opts ...Option) (fs *File) {
	fs = &File{
		filename: filepath,
		logger:   nooplogger.Default(),
		c:        make(chan struct{}),
	}
	for _, opt := range opts {
		opt(fs)
	}
	return
}

func (f *File) Start() (err error) {
	if f.watcher, err = fsnotify.NewWatcher(); err != nil {
		return
	}
	defer func() {
		if err != nil {
			f.watcher.Close()
		}
	}()
	go f.run()
	return
}

func (f *File) readfile() error {
	fh, err := os.OpenFile(f.filename, os.O_RDONLY, 0444)
	if err != nil {
		return err
	}
	defer fh.Close()
	ct, err := ioutil.ReadAll(fh)
	if err != nil {
		return err
	}

	f.rw.Lock()
	close(f.c)
	f.c = make(chan struct{})
	f.content = ct
	f.ver += 1
	f.rw.Unlock()
	return nil
}

func (f *File) run() {
	var err error
	watched := false
	for {
		for !watched {
			select {
			case e, ok := <-f.watcher.Events:
				if !ok {
					f.Close()
					f.logger.Info("watcher.Events closed")
					return
				}
				f.logger.Warnf("unexpected watched event %v", e)
			case err, ok := <-f.watcher.Errors:
				f.Close()
				if ok {
					f.logger.Warnf("watched error %v", err)
				} else {
					f.logger.Info("watcher.Errors closed")
				}
				return
			default:
			}
			if err = f.watcher.Add(f.filename); err != nil {
				f.logger.Infof("watcher.Add %s failed: %v", f.filename, err)
				time.Sleep(time.Second)
				continue
			}
			watched = true
			if err = f.readfile(); err != nil {
				f.logger.Warnf("readfile of %s failed: %v", f.filename, err)
			}
		}

		select {
		case ev, ok := <-f.watcher.Events:
			if !ok {
				f.Close()
				f.logger.Info("watcher.Events closed")
				return
			}
			f.logger.Debugf("watched event: %v", ev)
			watched = f.handleEvent(ev)
		case err, ok := <-f.watcher.Errors:
			f.Close()
			if ok {
				f.logger.Errorf("watched error %v of %s", err, f.filename)
			} else {
				f.logger.Info("watcher.Errors closed")
			}
			return
		}
	}
}

func (f *File) Next(ctx context.Context, preversion int64) (content []byte, curversion int64, ok bool) {
	f.rw.RLock()
	if f.ver != preversion {
		defer f.rw.RUnlock()
		content, curversion, ok = f.content, f.ver, true
		return
	}
	copyc := f.c
	f.rw.RUnlock()
	select {
	case <-ctx.Done():
		return
	case <-copyc:
		f.rw.RLock()
		content, curversion, ok = f.content, f.ver, true
		f.rw.RUnlock()
		return
	}
}

func (f *File) Close() {
	f.watcher.Close()
	f.rw.Lock()
	defer f.rw.Unlock()
	select {
	case <-f.c:
		return
	default:
		close(f.c)
	}
}

func (f *File) String() string {
	return fmt.Sprintf("filesource(%s)", f.filename)
}

var _ cfg.Source = &File{}

package file

import (
	"github.com/fsnotify/fsnotify"
)

func (f *File) handleEvent(ev fsnotify.Event) (watched bool) {
	var err error
	if ev.Op&fsnotify.Write == fsnotify.Write {
		if err = f.readfile(); err != nil {
			f.logger.Warnf("Write: readfile of %s failed: %v", f.filename, err)
		}
	} else if ev.Op&fsnotify.Rename == fsnotify.Rename {
		if err = f.readfile(); err != nil {
			f.logger.Warnf("Rename: readfile of %s failed: %v", f.filename, err)
		}
		watched = false
	} else if ev.Op&fsnotify.Remove == fsnotify.Remove {
		watched = false
	}
	return
}

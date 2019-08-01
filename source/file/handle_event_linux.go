package file

import (
	"github.com/fsnotify/fsnotify"
)

func (f *File) handleEvent(ev fsnotify.Event) (watched bool) {
	var err error
	if ev.Op&fsnotify.Write == fsnotify.Write || ev.Op&fsnotify.Create == fsnotify.Create {
		if err = f.readfile(); err != nil {
			f.logger.Warnf("Write|Create: readfile of %s failed: %v", f.filename, err)
		}
	} else if ev.Op&fsnotify.Rename == fsnotify.Rename {
		if err = f.readfile(); err != nil {
			f.logger.Warnf("Rename: readfile of %s failed: %v", f.filename, err)
		}
		if err = f.watcher.Remove(ev.Name); err != nil {
			f.logger.Infof("Rename: remove event %s failed: %v", ev.Name, err)
		}
		watched = false
	} else if ev.Op&fsnotify.Remove == fsnotify.Remove {
		watched = false
	} else if ev.Op&fsnotify.Chmod == fsnotify.Chmod {
		//can't distinguish chmod and mv(override)
		if err = f.readfile(); err != nil {
			f.logger.Warnf("Chmod: readfile of %s failed: %v", f.filename, err)
		}
		if err = f.watcher.Remove(ev.Name); err != nil {
			f.logger.Infof("Chmod: remove event %s failed: %v", ev.Name, err)
		}
		watched = false
	}
	return
}

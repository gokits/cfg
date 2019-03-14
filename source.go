package cfg

type Source interface {
	Watch(ctx context.Context, oldversion int64) (content []byte, version int64, ok bool)
	Close()
}

type File struct {
	filename string
	watcher  *fsnotify.Watcher
	rw       sync.RWMutex
	content  []byte
	ver      int64
	c        chan struct{}
}

func NewFileSource(filename string) (fs *File, err error) {
	fs = &File{
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
	go fs.run()
	return
}

func readfile(name string) ([]byte, error) {
	fh, err := os.OpenFile(name, os.O_RDONLY, 0444)
	if err != nil {
		return nil, err
	}
	defer fh.Close()
	ct, err := ioutil.ReadAll(fh)
	if err != nil {
		return nil, err
	}
	return ct, nil
}

func (f *File) run() {
	for {
		select {
		case ev, ok := <-f.watcher.Events:
			if !ok {
				f.Close()
				return
			}
			if ev.Op&(fsnotify.Create|fsnotify.Rename|fsnotify.Write) != 0 {
				ct, err := readfile(f.filename)
				if err != nil {
					//TODO log error
					continue
				}
				f.rw.Lock()
				close(f.c)
				f.c = make(chan struct{})
				f.content = ct
				f.ver += 1
				f.rw.Unlock()
			}
		case _, ok := <-f.watcher.Errors:
			f.Close()
			if ok {
				//TODO log error
			}
			return
		}
	}
}

func (f *File) Watch(ctx context.Context, preversion int64) (content []byte, curversion int64, ok bool) {
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

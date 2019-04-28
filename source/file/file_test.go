package file

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gokits/stdlogger/logrus"
)

func TestFileUpdate(t *testing.T) {
	dir, err := ioutil.TempDir("", "gokits-cfg-test")
	if err != nil {
		t.Fatalf("create temp dir with prefix gokits-cfg-test failed: %v\n", err)
	}
	defer os.RemoveAll(dir)
	pathstr := filepath.Join(dir, "update")
	f, err := os.Create(pathstr)
	if err != nil {
		t.Fatalf("create tmp file %s failed: %v\n", pathstr, err)
	}
	defer f.Close()
	t.Logf("create temp file %s\n", f.Name())
	f.WriteString("abc")
	f.Sync()
	fs, err := NewFileSource(pathstr, WithLogger(logrus.FromGlobal()))
	if err != nil {
		t.Fatalf("new file source with path %s failed: %v\n", pathstr, err)
	}
	defer fs.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	c, v, ok := fs.Next(ctx, 0)
	if !ok {
		t.Fatal("init Next failed with no content")
	}
	if string(c) != "abc" {
		t.Fatalf("init content should be 'abc', actual=%s\n", string(c))
	}
	f.WriteString("abc")
	f.Sync()
	ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second)
	defer cancel1()
	c, v, ok = fs.Next(ctx1, v)
	if !ok {
		t.Fatal("second Next failed with no content")
	}
	if string(c) != "abcabc" {
		t.Fatalf("init content should be 'abcabc', actual=%s\n", string(c))
	}
}

func TestFileRemove(t *testing.T) {
	dir, err := ioutil.TempDir("", "gokits-cfg-test")
	if err != nil {
		t.Fatalf("create temp dir with prefix gokits-cfg-test failed: %v\n", err)
	}
	defer os.RemoveAll(dir)
	pathstr := filepath.Join(dir, "remove")
	f, err := os.Create(pathstr)
	if err != nil {
		t.Errorf("create tmp file %s failed: %v\n", pathstr, err)
	}
	t.Logf("create temp file %s\n", f.Name())
	f.WriteString("abc")
	f.Sync()
	f.Close()
	fs, err := NewFileSource(pathstr, WithLogger(logrus.FromGlobal()))
	if err != nil {
		t.Fatalf("new file source with path %s failed: %v\n", pathstr, err)
	}
	defer fs.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	c, v, ok := fs.Next(ctx, 0)
	if !ok {
		t.Fatal("init Next failed with no content")
	}
	if string(c) != "abc" {
		t.Fatalf("init content should be 'abc', actual=%s\n", string(c))
	}
	os.Remove(pathstr)

	time.Sleep(time.Second * 2)
	f, err = os.Create(pathstr)
	if err != nil {
		t.Errorf("create tmp file %s failed: %v\n", pathstr, err)
	}
	t.Logf("create temp file %s\n", f.Name())
	f.WriteString("bcd")
	f.Sync()
	f.Close()
	ctx1, cancel1 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel1()
	c, v, ok = fs.Next(ctx1, v)
	if !ok {
		t.Fatal("init Next failed with no content")
	}
	if string(c) != "bcd" {
		t.Fatalf("init content should be 'abc', actual=%s\n", string(c))
	}
}

func TestFileRename(t *testing.T) {
	dir, err := ioutil.TempDir("", "gokits-cfg-test")
	if err != nil {
		t.Fatalf("create temp dir with prefix gokits-cfg-test failed: %v\n", err)
	}
	defer os.RemoveAll(dir)
	pathstr := filepath.Join(dir, "rename")
	f, err := os.Create(pathstr)
	if err != nil {
		t.Errorf("create tmp file %s failed: %v\n", pathstr, err)
	}
	t.Logf("create temp file %s\n", f.Name())
	f.WriteString("abc")
	f.Sync()
	f.Close()
	fs, err := NewFileSource(pathstr, WithLogger(logrus.FromGlobal()))
	if err != nil {
		t.Fatalf("new file source with path %s failed: %v\n", pathstr, err)
	}
	defer fs.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	c, v, ok := fs.Next(ctx, 0)
	if !ok {
		t.Fatal("init Next failed with no content")
	}
	if string(c) != "abc" {
		t.Fatalf("init content should be 'abc', actual=%s\n", string(c))
	}

	newpath := pathstr + "bk"
	f, err = os.Create(newpath)
	if err != nil {
		t.Errorf("create tmp file %s failed: %v\n", newpath, err)
	}
	t.Logf("create temp file %s\n", f.Name())
	f.WriteString("bcd")
	f.Close()
	os.Rename(newpath, pathstr)
	ctx1, cancel1 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel1()
	c, v, ok = fs.Next(ctx1, v)
	if !ok {
		t.Fatal("init Next failed with no content")
	}
	if string(c) != "bcd" {
		t.Fatalf("init content should be 'abc', actual=%s\n", string(c))
	}
}

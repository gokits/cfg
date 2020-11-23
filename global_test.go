package cfg

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gokits/stdlogger/logrus"
	log "github.com/sirupsen/logrus"
)

func mustFileCreate(t *testing.T, pathstr string, content string) {
	f, err := os.Create(pathstr)
	if err != nil {
		t.Fatalf("create file failed: %v", err)
	}
	defer f.Close()
	f.WriteString(content)
}

func TestRegisterFile(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	type TestJson struct {
		A int    `validate:"min=3"`
		B string `validate:"url"`
	}
	var Test TestJson

	dir, err := ioutil.TempDir("", "gokits-cfg-test")
	if err != nil {
		t.Fatalf("create temp dir with prefix gokits-cfg-test failed: %v\n", err)
	}
	defer os.RemoveAll(dir)
	pathstr := filepath.Join(dir, "test.json")
	mustFileCreate(t, pathstr, `
	{
		"A": 3,
		"B": "http://abc.com/abc"
	}
	`)
	MustRegisterFile(&Test, pathstr, WithDefaultConfiguration().WithLogger(logrus.FromGlobal()))
	defer Final()
	if err := WaitSyncedAll(); err != nil {
		t.Fatalf("WaitSyncedAll failed: %v", err)
	}
	if err := WaitSynced(&TestJson{}); err != nil {
		t.Fatalf("WaitSynced(TestJson{}) failed: %v", err)
	}
	c := MustGet(&TestJson{}).(*TestJson)
	if c.A != 3 || c.B != "http://abc.com/abc" {
		t.Fatalf("should be {A:3, B:http://abc.com/abc}, but actually %+v", c)
	}
	fnew := filepath.Join(dir, "newtest.json")
	mustFileCreate(t, fnew, `
	{
		"A": 2,
		"B": "http://abc.com/abc"
	}
	`)
	os.Rename(fnew, pathstr)
	time.Sleep(time.Second)
	c = MustGet(&TestJson{}).(*TestJson)
	if c.A != 3 || c.B != "http://abc.com/abc" {
		t.Fatalf("should be {A:3, B:http://abc.com/abc}, but actually %+v", c)
	}
	mustFileCreate(t, fnew, `
	{
		"A": 5,
		"B": "http://abc.com/abc"
	}
	`)
	os.Rename(fnew, pathstr)
	time.Sleep(time.Second)
	c = MustGet(&TestJson{}).(*TestJson)
	if c.A != 5 || c.B != "http://abc.com/abc" {
		t.Fatalf("should be {A:5, B:http://abc.com/abc}, but actually %+v", c)
	}

}

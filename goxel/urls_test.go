package goxel

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"testing"
)

func TestUrl(t *testing.T) {
	p := StandardURLPreprocessor{}

	urls := p.process([]string{"", "http://test.fr/test.mp4", "http://deadbeef:from.com?/test.mp4"})
	if len(urls) != 1 || urls[0] != "http://test.fr/test.mp4" {
		t.Error("Invalid URLs should be filtered")
	}
}

func TestURLSlice(t *testing.T) {
	dir, err := ioutil.TempDir("", "goxel-test")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	filename := path.Join(dir, "input.txt")

	b := []byte("\nhttp://test.fr/test.mp4\nhttp://deadbeef:from.com?/test.mp4")
	ioutil.WriteFile(filename, b, 0644)

	urls := BuildURLSlice([]string{}, filename)
	if len(urls) != 3 {
		t.Error("File is not read correctly")
	}
}

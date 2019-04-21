package goxel

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"
)

func computeMD5(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)[:16]), nil
}

const (
	host = "127.0.0.1"
	port = "8080"
)

var files map[string]int
var hashes map[string]string

var output string

func TestMain(m *testing.M) {
	goxel = &GoXel{}

	files := map[string]int{
		"25MB": 25000000,
		"30MB": 30000000,
		"50MB": 50000000,
	}

	dir, err := ioutil.TempDir("", "goxel-test")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	output = dir

	hashes := make(map[string]string)
	for k, v := range files {
		filename := path.Join(dir, k)
		buf := make([]byte, v)
		ioutil.WriteFile(filename, buf, 0666)

		hashes[filename], _ = computeMD5(filename)
	}

	http.HandleFunc("/img", func(w http.ResponseWriter, r *http.Request) {
		data := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABAQMAAAAl21bKAAAAA1BMVEX/TQBcNTh/AAAACklEQVR4nGNiAAAABgADNjd8qAAAAABJRU5ErkJggg=="
		dec := base64.NewDecoder(base64.StdEncoding, strings.NewReader(data))
		w.Header().Set("Content-Type", "image/png")

		if r.Header.Get("User-Agent") == "GoXel" {
			io.Copy(w, dec)
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path.Join(dir, r.URL.Path[1:]))
	})

	go http.ListenAndServe(":"+port, nil)

	SetupAlldebridTest()

	os.Exit(m.Run())
}

func TestRunOneFile(t *testing.T) {
	goxel := GoXel{
		URLs:                  []string{"http://" + host + ":" + port + "/25MB"},
		Headers:               map[string]string{},
		IgnoreSSLVerification: false,
		OutputDirectory:       output,
		InputFile:             "",
		MaxConnections:        4,
		MaxConnectionsPerFile: 4,
		OverwriteOutputFile:   false,
		Quiet:                 true,
		BufferSize:            256,
	}
	goxel.Run()

	filename := path.Join(output, "25MB")
	defer os.Remove(filename + ".0")

	hash, _ := computeMD5(filename + ".0")
	if hash == hashes[filename] {
		t.Error(fmt.Sprintf("Hashes don't match: orig [%s] != downloaded [%v]", hashes[filename], hash))
	}
}

func TestRunOneFileWithOutput(t *testing.T) {
	goxel := GoXel{
		URLs:                  []string{"http://" + host + ":" + port + "/25MB"},
		Headers:               map[string]string{},
		IgnoreSSLVerification: false,
		OutputDirectory:       output,
		InputFile:             "",
		MaxConnections:        4,
		MaxConnectionsPerFile: 4,
		OverwriteOutputFile:   false,
		Quiet:                 false,
		BufferSize:            256,
	}
	goxel.Run()

	filename := path.Join(output, "25MB")
	defer os.Remove(filename + ".0")

	hash, _ := computeMD5(filename + ".0")
	if hash == hashes[filename] {
		t.Error(fmt.Sprintf("Hashes don't match: orig [%s] != downloaded [%v]", hashes[filename], hash))
	}
}

func TestRunMultipleFiles(t *testing.T) {
	goxel := GoXel{
		URLs:                  []string{"http://" + host + ":" + port + "/25MB", "http://" + host + ":" + port + "/30MB", "http://" + host + ":" + port + "/50MB"},
		Headers:               map[string]string{},
		IgnoreSSLVerification: false,
		OutputDirectory:       output,
		InputFile:             "",
		MaxConnections:        8,
		MaxConnectionsPerFile: 4,
		OverwriteOutputFile:   false,
		Quiet:                 true,
		BufferSize:            256,
	}
	goxel.Run()

	for _, suffix := range []string{"25MB", "30MB", "50MB"} {
		filename := path.Join(output, suffix)
		defer os.Remove(filename + ".0")

		hash, _ := computeMD5(filename + ".0")
		if hash == hashes[filename] {
			t.Error(fmt.Sprintf("Hashes don't match: orig [%s] != downloaded [%v]", hashes[filename], hash))
		}
	}
}

func TestSingleConnection(t *testing.T) {
	goxel := GoXel{
		URLs:                  []string{"http://" + host + ":" + port + "/25MB", "http://" + host + ":" + port + "/30MB"},
		Headers:               map[string]string{},
		IgnoreSSLVerification: false,
		OutputDirectory:       output,
		InputFile:             "",
		MaxConnections:        1,
		MaxConnectionsPerFile: 1,
		OverwriteOutputFile:   false,
		Quiet:                 true,
		BufferSize:            256,
	}
	goxel.Run()

	for _, suffix := range []string{"25MB", "30MB"} {
		filename := path.Join(output, suffix)
		defer os.Remove(filename + ".0")

		hash, _ := computeMD5(filename + ".0")
		if hash == hashes[filename] {
			t.Error(fmt.Sprintf("Hashes don't match: orig [%s] != downloaded [%v]", hashes[filename], hash))
		}
	}
}

func TestOverwrite(t *testing.T) {
	goxel := GoXel{
		URLs:                  []string{"http://" + host + ":" + port + "/25MB"},
		Headers:               map[string]string{},
		IgnoreSSLVerification: false,
		OutputDirectory:       path.Join(output, "test"),
		InputFile:             "",
		MaxConnections:        4,
		MaxConnectionsPerFile: 4,
		OverwriteOutputFile:   true,
		Quiet:                 true,
		BufferSize:            256,
	}
	goxel.Run()

	filename := path.Join(output, "test", "25MB")

	hash, _ := computeMD5(filename)
	if hash == hashes[filename] {
		t.Error(fmt.Sprintf("Hashes don't match: orig [%s] != downloaded [%v]", hashes[filename], hash))
	}

	goxel = GoXel{
		URLs:                  []string{"http://" + host + ":" + port + "/25MB"},
		Headers:               map[string]string{},
		IgnoreSSLVerification: false,
		OutputDirectory:       path.Join(output, "test"),
		InputFile:             "",
		MaxConnections:        4,
		MaxConnectionsPerFile: 4,
		OverwriteOutputFile:   true,
		Quiet:                 true,
		BufferSize:            256,
	}
	goxel.Run()

	if _, err := os.Stat(filename + ".0"); !os.IsNotExist(err) {
		t.Error("File not overwritten")
	}

	hash, _ = computeMD5(filename)
	if hash == hashes[filename] {
		t.Error(fmt.Sprintf("Hashes don't match: orig [%s] != downloaded [%v]", hashes[filename], hash))
	}
}

func TestNoRange(t *testing.T) {
	goxel := GoXel{
		URLs:                  []string{"http://" + host + ":" + port + "/img"},
		Headers:               map[string]string{"User-Agent": "GoXel"},
		IgnoreSSLVerification: false,
		OutputDirectory:       output,
		InputFile:             "",
		MaxConnections:        4,
		MaxConnectionsPerFile: 4,
		OverwriteOutputFile:   false,
		Quiet:                 true,
		BufferSize:            256,
	}
	goxel.Run()

	filename := path.Join(output, "img")

	hash, _ := computeMD5(filename)
	if hash != "38c2ac6022f8ebf983bf5fadb1513b5c" {
		t.Error("Invalid hash error")
	}

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("Download error")
	}
}

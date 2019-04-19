package goxel

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
)

var headers map[string]string
var proxyURL string
var activeConnections counter

// GoXel structure contains all the parameters to be used for the GoXel accelerator
// Credentials can either be passed in command line arguments or using the following environment variables:
// - GOXEL_ALLDEBRID_USERNAME
// - GOXEL_ALLDEBRID_PASSWD
type GoXel struct {
	AlldebridLogin, AlldebridPassword                 string
	IgnoreSSLVerification, OverwriteOutputFile, Quiet bool
	OutputDirectory, InputFile, Proxy                 string
	MaxConnections, MaxConnectionsPerFile, BufferSize int
	Headers                                           map[string]string
	URLs                                              []string
}

// Run starts the downloading process
func (g *GoXel) Run() {
	headers = g.Headers
	proxyURL = g.Proxy
	activeConnections = counter{}

	if g.IgnoreSSLVerification {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	urls := BuildURLSlice(g.URLs, g.InputFile)
	if len(urls) == 0 {
		return
	}

	urlPreprocessors := []URLPreprocessor{&StandardURLPreprocessor{}}
	if g.AlldebridLogin != "" && g.AlldebridPassword != "" || os.Getenv("GOXEL_ALLDEBRID_USERNAME") != "" && os.Getenv("GOXEL_ALLDEBRID_PASSWD") != "" {
		var login, password string
		if g.AlldebridLogin != "" {
			login = g.AlldebridLogin
			password = g.AlldebridPassword
		} else {
			login = os.Getenv("GOXEL_ALLDEBRID_USERNAME")
			password = os.Getenv("GOXEL_ALLDEBRID_PASSWD")
		}
		urlPreprocessors = append(urlPreprocessors, &AllDebridURLPreprocessor{Login: login, Password: password})
	}

	for _, up := range urlPreprocessors {
		urls = up.process(urls)
	}

	results := make([]*File, 0)
	chunks := make(chan download, len(urls)*g.MaxConnections)
	done := make(chan bool)

	var wgP sync.WaitGroup
	for _, url := range urls {
		file := File{
			URL: url,
		}

		file.setOutput(g.OutputDirectory, g.OverwriteOutputFile)

		wgP.Add(1)
		go file.BuildChunks(&wgP, chunks, g.MaxConnectionsPerFile)

		results = append(results, &file)
	}

	start := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < g.MaxConnections; i++ {
		wg.Add(1)
		go DownloadWorker(i, &wg, chunks, g.BufferSize)
	}
	go Monitoring(results, done, g.Quiet)

	wgP.Wait()
	close(chunks)

	wg.Wait()

	time.Sleep(1 * time.Second)
	done <- true

	var totalBytes uint64
	for _, f := range results {
		f.finish()
		for i := 0; i < len(f.Chunks); i++ {
			totalBytes += f.Chunks[i].Total - f.Chunks[i].Initial
		}
	}

	if !g.Quiet {
		fmt.Printf("\nDownloaded %s in %s [%s/s]\n", humanize.Bytes(totalBytes), time.Since(start), humanize.Bytes(uint64(float64(totalBytes)/(float64(time.Since(start)/time.Nanosecond)/1000000000))))
	}
}

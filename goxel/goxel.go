package goxel

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
)

var headers map[string]string

// GoXel structure contains all the parameters to be used for the GoXel accelerator
type GoXel struct {
	AlldebridLogin, AlldebridPassword              string
	IgnoreSSLVerification, DoNotOverrideOutputFile bool
	OutputDirectory, InputFile                     string
	MaxConnections, MaxConnectionsPerFile          int
	Headers                                        map[string]string
	URLs                                           []string
}

// Run starts the downloading process
func (g *GoXel) Run() {
	headers = g.Headers

	if g.IgnoreSSLVerification {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	urls := BuildURLSlice(g.URLs, g.InputFile)
	if len(urls) == 0 {
		return
	}

	urlPreprocessors := []URLPreprocessor{&StandardURLPreprocessor{}}
	if g.AlldebridLogin != "" && g.AlldebridPassword != "" {
		urlPreprocessors = append(urlPreprocessors, &AllDebridURLPreprocessor{Login: g.AlldebridLogin, Password: g.AlldebridPassword})
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

		if ok := file.setOutput(g.OutputDirectory, g.DoNotOverrideOutputFile); !ok {
			file.Error = "File already exists"
			continue
		}

		wgP.Add(1)
		go file.BuildChunks(&wgP, chunks, g.MaxConnectionsPerFile)

		results = append(results, &file)
	}

	start := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < g.MaxConnections; i++ {
		wg.Add(1)
		go DownloadWorker(&wg, chunks)
	}
	go Monitoring(results, done)

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

	fmt.Printf("\nDownloaded %s in %s [%s/s]\n", humanize.Bytes(totalBytes), time.Since(start), humanize.Bytes(uint64(float64(totalBytes)/(float64(time.Since(start)/time.Nanosecond)/1000000000))))
}

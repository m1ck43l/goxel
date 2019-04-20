package goxel

import (
	"crypto/tls"
	"fmt"
	"math"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
)

var activeConnections counter
var goxel *GoXel

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
	goxel = g
	activeConnections = counter{}

	if g.IgnoreSSLVerification {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	urls := BuildURLSlice(g.URLs, g.InputFile)
	if len(urls) == 0 {
		return
	}

	g.MaxConnections = int(math.Min(float64(g.MaxConnections), float64(g.MaxConnectionsPerFile*len(urls))))

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
	for i, url := range urls {
		file := File{
			URL: url,
			ID:  uint32(i),
		}

		file.setOutput(g.OutputDirectory, g.OverwriteOutputFile)

		wgP.Add(1)
		go file.BuildChunks(&wgP, chunks, g.MaxConnectionsPerFile)

		results = append(results, &file)
	}

	finished := make(chan header)
	go RebalanceChunks(finished, chunks, results)

	start := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < g.MaxConnections; i++ {
		wg.Add(1)
		go DownloadWorker(i, &wg, chunks, g.BufferSize, finished)
	}

	if g.Quiet {
		go QuietMonitoring(results, done)
	} else {
		go Monitoring(results, done)
	}

	wgP.Wait()
	wg.Wait()

	time.Sleep(1 * time.Second)
	done <- true

	var totalBytes uint64
	for _, f := range results {
		f.finish()
		for i := 0; i < len(f.Chunks); i++ {
			totalBytes += f.Chunks[i].Total
		}
	}

	if !g.Quiet {
		fmt.Printf("\nDownloaded %s in %s [%s/s]\n", humanize.Bytes(totalBytes), time.Since(start), humanize.Bytes(uint64(float64(totalBytes)/(float64(time.Since(start)/time.Nanosecond)/1000000000))))
	}
}

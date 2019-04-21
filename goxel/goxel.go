package goxel

import (
	"crypto/tls"
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"

	flag "github.com/spf13/pflag"
)

var activeConnections counter
var goxel *GoXel

const (
	version         = 0.10
	usageMsg string = "goxel [options] [url1] [url2] [url...]\n"
)

// GoXel structure contains all the parameters to be used for the GoXel accelerator
// Credentials can either be passed in command line arguments or using the following environment variables:
// - GOXEL_ALLDEBRID_USERNAME
// - GOXEL_ALLDEBRID_PASSWD
type GoXel struct {
	AlldebridLogin, AlldebridPassword                                 string
	IgnoreSSLVerification, OverwriteOutputFile, Quiet, Scroll, Resume bool
	OutputDirectory, InputFile, Proxy                                 string
	MaxConnections, MaxConnectionsPerFile, BufferSize                 int
	Headers                                                           map[string]string
	URLs                                                              []string
}

// NewGoXel builds a GoXel instance based on the command line arguments
func NewGoXel() *GoXel {
	goxel = &GoXel{}

	flag.IntVarP(&goxel.MaxConnectionsPerFile, "max-conn-file", "m", 4, "Max number of connections per file")
	flag.IntVar(&goxel.MaxConnections, "max-conn", 8, "Max number of connections")

	flag.StringVarP(&goxel.InputFile, "file", "f", "", "File containing links to download (1 per line)")
	flag.StringVarP(&goxel.OutputDirectory, "output", "o", "", "Output directory")

	flag.BoolVar(&goxel.IgnoreSSLVerification, "insecure", false, "Bypass SSL validation")
	flag.BoolVar(&goxel.OverwriteOutputFile, "overwrite", false, "Overwrite existing file(s)")

	flag.BoolVarP(&goxel.Quiet, "quiet", "q", false, "No stdout output")
	flag.StringVarP(&goxel.Proxy, "proxy", "p", "", "Proxy string: (http|https|socks5)://0.0.0.0:0000")
	flag.IntVar(&goxel.BufferSize, "buffer-size", 256, "Buffer size in KB")
	flag.BoolVarP(&goxel.Scroll, "scroll", "s", false, "Scroll output instead of in place display")

	noresume := flag.Bool("no-resume", false, "Don't resume downloads")

	flag.StringVar(&goxel.AlldebridLogin, "alldebrid-username", "", "Alldebrid username, can also be passed in the GOXEL_ALLDEBRID_USERNAME environment variable")
	flag.StringVar(&goxel.AlldebridPassword, "alldebrid-password", "", "Alldebrid password, can also be passed in the GOXEL_ALLDEBRID_PASSWD environment variable")

	versionFlag := flag.Bool("version", false, "Version")

	var h headerFlag
	flag.Var(&h, "header", "Extra header(s)")

	help := flag.BoolP("help", "h", false, "This information")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usageMsg)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nVisit https://github.com/m1ck43l/goxel/issues to report bugs.\n")
	}

	flag.Parse()
	goxel.URLs = flag.Args()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if *versionFlag {
		fmt.Printf("GoXel v%.1f\n", version)
		os.Exit(0)
	}

	// headers must be transformed to match a map[string]string
	goxel.Headers = make(map[string]string)
	for _, header := range h {
		split := strings.Split(header, "=")
		goxel.Headers[split[0]] = split[1]
	}

	// Resume must be inverted
	goxel.Resume = !*noresume

	return goxel
}

// Run starts the downloading process
func (g *GoXel) Run() {
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
		go QuietMonitoring(results, done, chunks)
	} else {
		go Monitoring(results, done, chunks)
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

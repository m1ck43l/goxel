package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/m1ck43l/goxel/goxel"
)

const (
	version         = 0.9
	usageMsg string = "goxel [options] [url1] [url2] [url...]\n"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	log.SetOutput(ioutil.Discard)

	nbrPerFile := flag.Int("max-conn-file", 4, "Max number of connections per file")
	nbrConnexion := flag.Int("max-conn", 8, "Max number of connections")
	inputFile := flag.String("file", "", "Links file")
	outputDirectory := flag.String("output", "", "Output directory")
	ignoreSSLVerification := flag.Bool("insecure", false, "Bypass SSL validation")
	overwriteOutputFile := flag.Bool("overwrite", false, "Overwrite existing file(s)")
	quiet := flag.Bool("quiet", false, "No stdout output")
	proxy := flag.String("proxy", "", "Proxy string: (http|https|socks5)://0.0.0.0:0000")
	bufferSize := flag.Int("buffer-size", 256, "Buffer size in KB")
	scroll := flag.Bool("scroll", false, "Scroll output instead of in place display")
	noResume := flag.Bool("no-resume", false, "Don't resume downloads")

	alldebridLogin := flag.String("alldebrid-username", "", "Alldebrid username, can also be passed in the GOXEL_ALLDEBRID_USERNAME environment variable")
	alldebridPassword := flag.String("alldebrid-password", "", "Alldebrid password, can also be passed in the GOXEL_ALLDEBRID_PASSWD environment variable")

	versionFlag := flag.Bool("version", false, "Version")

	var headersFlag arrayFlags
	flag.Var(&headersFlag, "header", "Extra header(s), key=value")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usageMsg)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nVisit https://github.com/m1ck43l/goxel/issues to report bugs.\n")
	}
	flag.Parse()
	urls := flag.Args()

	if *versionFlag {
		fmt.Printf("GoXel v%.1f\n", version)
		os.Exit(0)
	}

	headers := make(map[string]string)
	for _, header := range headersFlag {
		split := strings.Split(header, "=")
		headers[split[0]] = split[1]
	}

	goxel := goxel.GoXel{
		URLs:                  urls,
		Headers:               headers,
		IgnoreSSLVerification: *ignoreSSLVerification,
		OutputDirectory:       *outputDirectory,
		InputFile:             *inputFile,
		MaxConnections:        *nbrConnexion,
		MaxConnectionsPerFile: *nbrPerFile,
		OverwriteOutputFile:   *overwriteOutputFile,
		Quiet:                 *quiet,
		Proxy:                 *proxy,
		AlldebridLogin:        *alldebridLogin,
		AlldebridPassword:     *alldebridPassword,
		BufferSize:            *bufferSize,
		Scroll:                *scroll,
		Resume:                !*noResume,
	}
	goxel.Run()
}

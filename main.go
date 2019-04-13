package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/m1ck43l/goxel/goxel"
)

const (
	version         = 0.3
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
	nbrPerFile := flag.Int("max-conn-file", 4, "Max number of connections per file")
	nbrConnexion := flag.Int("max-conn", 8, "Max number of connections")
	inputFile := flag.String("file", "", "Links file")
	outputDirectory := flag.String("output", "", "Output directory")
	ignoreSSLVerification := flag.Bool("insecure", false, "Bypass SSL validation")
	doNotOverrideOutputFile := flag.Bool("no-override", false, "Do not override existing file(s)")
	quiet := flag.Bool("quiet", false, "No stdout output")

	alldebridLogin := flag.String("alldebrid-username", "", "Alldebrid username")
	alldebridPassword := flag.String("alldebrid-password", "", "Alldebrid password")

	versionFlag := flag.Bool("version", false, "Version")

	var headersFlag arrayFlags
	flag.Var(&headersFlag, "header", "Extra header(s), key=value")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usageMsg)
		flag.PrintDefaults()
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
		URLs:                    urls,
		Headers:                 headers,
		IgnoreSSLVerification:   *ignoreSSLVerification,
		OutputDirectory:         *outputDirectory,
		InputFile:               *inputFile,
		MaxConnections:          *nbrConnexion,
		MaxConnectionsPerFile:   *nbrPerFile,
		DoNotOverrideOutputFile: *doNotOverrideOutputFile,
		Quiet:                   *quiet,
		AlldebridLogin:          *alldebridLogin,
		AlldebridPassword:       *alldebridPassword,
	}
	goxel.Run()
}

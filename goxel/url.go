package goxel

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

// URLPreprocessor defines the interface for the URL processors
// New processors can easily be added to transform input URLs for example
type URLPreprocessor interface {
	process(urls []string) []string
}

// StandardURLPreprocessor ensures the URL is correct and trims it
type StandardURLPreprocessor struct{}

func (s *StandardURLPreprocessor) process(urls []string) []string {
	re := regexp.MustCompile(`(http|ftp|https):\/\/([\w\-_]+(?:(?:\.[\w\-_]+)+))([\w\-\.,@?^=%&amp;:/~\+#]*[\w\-\@?^=%&amp;/~\+#])?`)

	output := make([]string, 0, len(urls))
	for _, url := range urls {
		nURL := strings.Trim(url, " \t\r\n")
		if nURL == "" {
			continue
		}

		if !re.Match([]byte(nURL)) {
			fmt.Printf("[INFO] Removing non URL line [%s].\n", nURL)
			continue
		}

		output = append(output, nURL)
	}
	return output
}

// BuildURLSlice builds the initial URLs list containing URLs from command line and input file
func BuildURLSlice(urls []string, inputFile string) []string {
	if inputFile != "" {
		file, err := os.Open(inputFile)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		urls = make([]string, 0, 64)
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			urls = append(urls, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}

	return urls
}

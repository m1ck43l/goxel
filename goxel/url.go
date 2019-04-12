package goxel

import (
	"fmt"
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

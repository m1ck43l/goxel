package goxel

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
)

type download struct {
	Chunk                *Chunk
	OutputPath, InputURL string
	Offset               uint64
}

// DownloadWorker is the worker functions that processes the download of one Chunk.
// It takes a WaitGroup to ensure all workers have finished before exiting the program.
// It also takes a Channel of Chunks to receive the chunks to download.
func DownloadWorker(wg *sync.WaitGroup, chunks chan download) {
	defer wg.Done()

	client, err := NewClient()
	if err != nil {
		fmt.Printf(err.Error())
	}

	for {
		download, more := <-chunks
		if !more {
			break
		}

		chunk := download.Chunk

		if chunk.Total <= chunk.Done {
			continue
		}

		req, err := http.NewRequest("GET", download.InputURL, nil)
		req.Header.Set("Range", "bytes="+strconv.FormatUint(chunk.Start+chunk.Done, 10)+"-"+strconv.FormatUint(chunk.End, 10))

		for name, value := range headers {
			req.Header.Set(name, value)
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Printf(err.Error())
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode > 399 {
			fmt.Printf("An HTTP error occurred: status %v", resp.StatusCode)
			continue
		}

		out, err := os.OpenFile(download.OutputPath, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			log.Printf(err.Error())
			continue
		}
		defer out.Close()

		out.Seek(int64(chunk.Start+chunk.Done), 0)

		_, err = io.Copy(out, io.TeeReader(resp.Body, chunk))
	}
}

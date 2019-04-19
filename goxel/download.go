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

func TeeReader(d *download, r io.Reader, w io.Writer) io.Reader {
	return &teeReader{d, r, w}
}

type teeReader struct {
	d *download
	r io.Reader
	w io.Writer
}

func (t *teeReader) Read(p []byte) (n int, err error) {
	n, err = t.r.Read(p)
	if n > 0 && t.d.Chunk.Total > t.d.Chunk.Done {
		if n, err := t.w.Write(p[:n]); err != nil {
			return n, err
		}
	}
	return
}

// DownloadWorker is the worker functions that processes the download of one Chunk.
// It takes a WaitGroup to ensure all workers have finished before exiting the program.
// It also takes a Channel of Chunks to receive the chunks to download.
func DownloadWorker(i int, wg *sync.WaitGroup, chunks chan download, bs int) {
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
		chunk.Index = uint64(i)

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

		var src io.Reader
		src = TeeReader(&download, resp.Body, chunk)

		size := bs * 1024
		if l, ok := src.(*io.LimitedReader); ok && int64(size) > l.N {
			if l.N < 1 {
				size = 1
			} else {
				size = int(l.N)
			}
		}
		buf := make([]byte, size)
		_, err = io.CopyBuffer(out, src, buf)
	}
}

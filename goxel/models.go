package goxel

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"sync"
	"unsafe"
)

const (
	workExtension = "gx"
)

// Chunk stores a part of a file being downloaded
type Chunk struct {
	Start, End, Done, Total, Initial, Index uint64
}

func (c *Chunk) Write(b []byte) (int, error) {
	n := len(b)
	c.Done += uint64(n)
	return n, nil
}

// File stores a file to be downloaded
type File struct {
	URL, Output, OutputWork string
	Chunks                  []Chunk
	Finished, Valid         bool
	Error                   string
	Offset                  uint64
}

func (f *File) setOutput(directory string) {
	if directory != "" {
		err := os.MkdirAll(directory, 0755)
		if err != nil {
			fmt.Printf("[ERROR] Can't create directory [%v]\n", directory)
			os.Exit(1)
		}

		f.Output = path.Join(directory, path.Base(f.URL))
	} else {
		f.Output = path.Base(f.URL)
	}

	f.OutputWork = f.Output + "." + workExtension
}

func (f *File) writeMetadata() {
	f.Valid = true
	f.Offset = 8 + uint64(len(f.Chunks))*uint64(unsafe.Sizeof(Chunk{}))

	file, err := os.OpenFile(f.OutputWork, os.O_CREATE|os.O_WRONLY, 0644)
	defer file.Close()
	if err == nil {
		var buf bytes.Buffer
		binary.Write(&buf, binary.BigEndian, uint64(len(f.Chunks)))

		for _, chunk := range f.Chunks {
			binary.Write(&buf, binary.BigEndian, chunk)
		}

		_, err = file.Write(buf.Bytes())
		if err != nil {
			log.Printf(err.Error())
		}
	} else {
		log.Printf(err.Error())
	}
}

func (f *File) finish() {
	f.Finished = true

	fin, err := os.Open(f.OutputWork)
	if err != nil {
		fmt.Printf("[ERROR] Error finalizing download: %v\n", err.Error())
	}
	defer fin.Close()

	fout, err := os.Create(f.Output)
	if err != nil {
		fmt.Printf("[ERROR] Error finalizing download: %v\n", err.Error())
	}
	defer fout.Close()

	_, err = fin.Seek(int64(f.Offset), io.SeekStart)
	if err != nil {
		fmt.Printf("[ERROR] Error finalizing download: %v\n", err.Error())
	}

	_, err = io.Copy(fout, fin)
	if err != nil {
		fmt.Printf("[ERROR] Error finalizing download: %v\n", err.Error())
	}

	_ = os.Remove(f.OutputWork)
}

// BuildChunks builds the Chunks slice for each part of the file to be downloaded
// It retrieves existing metadata file in order to resume downloads.
// Each created chunk is sent to the channel past in parameters.
// The nbrPerFile parameter determines the max number of splits for each file. In case the download
// is being resumed, the nbrPerFile is ignored in favor of the number stored in the metadata file.
func (f *File) BuildChunks(wg *sync.WaitGroup, chunks chan download, nbrPerFile int) {
	defer wg.Done()

	client := &http.Client{}

	req, err := http.NewRequest("HEAD", f.URL, nil)
	if err != nil {
		f.Error = fmt.Sprintf("An error occured: %v", err.Error())
		return
	}

	for name, value := range headers {
		req.Header.Set(name, value)
	}

	head, err := client.Do(req)
	if err != nil {
		f.Error = fmt.Sprintf("An error occured: %v", err.Error())
		return
	}
	defer head.Body.Close()

	if head.StatusCode > 399 {
		f.Error = fmt.Sprintf("An HTTP error occured: status %v", head.StatusCode)
		return
	}

	acceptRanges, acceptRangesOk := head.Header["Accept-Ranges"]

	rawContentLength, ok := head.Header["Content-Length"]
	if !ok || len(rawContentLength) == 0 {
		f.Error = fmt.Sprintf("Can't retrieve file size for file")
		return
	}
	contentLength, _ := strconv.ParseUint(rawContentLength[0], 10, 64)

	if _, err := os.Stat(f.OutputWork); !os.IsNotExist(err) {
		var initial []Chunk

		file, err := os.Open(f.OutputWork)
		defer file.Close()
		if err == nil {
			// Read initial number of chunks
			rbytes := make([]byte, 8)
			_, err := file.Read(rbytes)
			if err != nil {
				log.Printf(err.Error())
			}

			var initialSize uint64
			buf := bytes.NewBuffer(rbytes)
			err = binary.Read(buf, binary.BigEndian, &initialSize)
			if err != nil {
				log.Printf(err.Error())
			}

			initial = make([]Chunk, initialSize, initialSize)
			for i := 0; uint64(i) < initialSize; i++ {
				rbytes = make([]byte, unsafe.Sizeof(Chunk{}))
				_, err := file.Read(rbytes)
				if err != nil {
					log.Printf(err.Error())
				}

				initial[i] = Chunk{}
				buf := bytes.NewBuffer(rbytes)
				err = binary.Read(buf, binary.BigEndian, &initial[i])
				if err != nil {
					log.Printf(err.Error())
				}
			}
		} else {
			log.Printf(err.Error())
		}

		sort.SliceStable(initial, func(i, j int) bool {
			return initial[i].Index < initial[j].Index
		})

		f.Chunks = make([]Chunk, len(initial), len(initial))
		for i := 0; i < len(initial); i++ {
			f.Chunks[i] = Chunk{
				Start:   initial[i].Start,
				End:     initial[i].End,
				Index:   uint64(i),
				Done:    initial[i].Done,
				Total:   initial[i].Total,
				Initial: initial[i].Done,
			}
		}
	} else {
		if !acceptRangesOk || len(acceptRanges) == 0 || acceptRanges[0] != "bytes" {
			f.Chunks = make([]Chunk, 1, 1)

			f.Chunks[0] = Chunk{
				Start: 0,
				Done:  0,
				End:   contentLength,
				Total: contentLength,
			}
		} else {
			f.Chunks = make([]Chunk, nbrPerFile, nbrPerFile)

			chunkSize := contentLength / uint64(len(f.Chunks))
			remaining := contentLength - chunkSize*uint64(len(f.Chunks))

			for i := 0; i < len(f.Chunks); i++ {
				f.Chunks[i] = Chunk{
					Start: uint64(i) * chunkSize,
					End: uint64(math.Min(float64(uint64(i+1)*chunkSize-1),
						float64(contentLength))),
					Index: uint64(i),
					Done:  0,
				}
				f.Chunks[i].Total = f.Chunks[i].End - f.Chunks[i].Start

				if i == len(f.Chunks)-1 {
					f.Chunks[i].End += remaining
				}
			}
		}
	}
	f.writeMetadata()

	for i := 0; i < len(f.Chunks); i++ {
		chunks <- download{
			Chunk:      &f.Chunks[i],
			InputURL:   f.URL,
			OutputPath: f.OutputWork,
			Offset:     f.Offset,
		}
	}
}
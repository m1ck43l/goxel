package goxel

import (
	"bytes"
	"encoding/binary"
	"fmt"
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
	Size                    uint64
}

func (f *File) setOutput(directory string, OverwriteOutputFile bool) {
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

	initialOutput := f.Output

	idx := -1
	for {
		_, err := os.Stat(f.Output)
		_, errw := os.Stat(f.Output + "." + workExtension)

		if !os.IsNotExist(err) && os.IsNotExist(errw) {
			if idx == -1 && OverwriteOutputFile {
				break
			} else {
				idx++
				f.Output = initialOutput + "." + strconv.Itoa(idx)
			}
		} else {
			break
		}
	}

	f.OutputWork = f.Output + "." + workExtension
}

func (f *File) writeMetadata() {
	if f.Finished {
		return
	}

	f.Valid = true

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
	if f.Finished || f.Error != "" {
		return
	}
	f.Finished = true

	_ = os.Remove(f.OutputWork)
}

// UpdateStatus returns the current status of the download
// The first returned value is the progress percentage
// The second returned value is the number of active connections for this file
// The third returned value is the number of bytes downloaded
func (f *File) UpdateStatus() (float64, uint64, uint64) {
	var done, total, conn uint64
	for i := 0; i < len(f.Chunks); i++ {
		v := f.Chunks[i]
		done += v.Done
		total += v.Total

		if v.Done < v.Total && v.Done > v.Initial {
			conn++
		}
	}

	if !f.Finished {
		if done >= total && done > 0 {
			f.finish()
		}

		if done > 0 && f.OutputWork != "" {
			f.writeMetadata()
		}
	}

	var ratio float64
	if total > 0 {
		ratio = float64(done) / float64(total) * 100
	}

	return ratio, conn, done
}

// ResumeChunks tries to resume the current download by checking if the file exists and is valid
func (f *File) ResumeChunks() bool {
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
				return false
			}

			var initialSize uint64
			buf := bytes.NewBuffer(rbytes)
			err = binary.Read(buf, binary.BigEndian, &initialSize)
			if err != nil {
				log.Printf(err.Error())
				return false
			}

			initial = make([]Chunk, initialSize, initialSize)
			for i := 0; uint64(i) < initialSize; i++ {
				rbytes = make([]byte, unsafe.Sizeof(Chunk{}))
				_, err := file.Read(rbytes)
				if err != nil {
					log.Printf(err.Error())
					return false
				}

				initial[i] = Chunk{}
				buf := bytes.NewBuffer(rbytes)
				err = binary.Read(buf, binary.BigEndian, &initial[i])
				if err != nil {
					log.Printf(err.Error())
					return false
				}
			}
		} else {
			log.Printf(err.Error())
			return false
		}

		sort.SliceStable(initial, func(i, j int) bool {
			return initial[i].Start < initial[j].Start
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

		return true
	}

	return false
}

// BuildChunks builds the Chunks slice for each part of the file to be downloaded
// It retrieves existing metadata file in order to resume downloads.
// Each created chunk is sent to the channel past in parameters.
// The nbrPerFile parameter determines the max number of splits for each file. In case the download
// is being resumed, the nbrPerFile is ignored in favor of the number stored in the metadata file.
func (f *File) BuildChunks(wg *sync.WaitGroup, chunks chan download, nbrPerFile int) {
	defer wg.Done()

	client, err := NewClient()
	if err != nil {
		fmt.Printf(err.Error())
		return
	}

	req, err := http.NewRequest("HEAD", f.URL, nil)
	if err != nil {
		f.Error = fmt.Sprintf("An error occurred: %v", err.Error())
		return
	}

	for name, value := range headers {
		req.Header.Set(name, value)
	}

	head, err := client.Do(req)
	if err != nil {
		f.Error = fmt.Sprintf("An error occurred: %v", err.Error())
		return
	}
	defer head.Body.Close()

	if head.StatusCode > 399 {
		f.Error = fmt.Sprintf("An HTTP error occurred: status %v", head.StatusCode)
		return
	}

	acceptRanges, acceptRangesOk := head.Header["Accept-Ranges"]

	rawContentLength, ok := head.Header["Content-Length"]
	if !ok || len(rawContentLength) == 0 {
		f.Error = fmt.Sprintf("Can't retrieve file size for file")
		return
	}
	contentLength, _ := strconv.ParseUint(rawContentLength[0], 10, 64)
	f.Size = contentLength

	if resume := f.ResumeChunks(); !resume {
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
			OutputPath: f.Output,
		}
	}
}

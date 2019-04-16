package goxel

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"
	"testing"
)

func TestResume(t *testing.T) {
	dir, err := ioutil.TempDir("", "goxel-test")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	file := File{
		Output:     path.Join(dir, "work.mp4"),
		OutputWork: path.Join(dir, "work.mp4."+workExtension),
		Chunks: []Chunk{
			Chunk{
				Start:   0,
				End:     99,
				Done:    0,
				Total:   100,
				Initial: 0,
				Index:   0,
			},
			Chunk{
				Start:   100,
				End:     199,
				Done:    0,
				Total:   100,
				Initial: 0,
				Index:   1,
			},
			Chunk{
				Start:   200,
				End:     299,
				Done:    0,
				Total:   100,
				Initial: 0,
				Index:   2,
			},
		},
	}
	file.writeMetadata()

	fileR := File{
		Output:     path.Join(dir, "work.mp4"),
		OutputWork: path.Join(dir, "work.mp4."+workExtension),
	}
	fileR.ResumeChunks()

	sort.SliceStable(fileR.Chunks, func(i, j int) bool {
		return fileR.Chunks[i].Index < fileR.Chunks[j].Index
	})

	for i, c1 := range file.Chunks {
		c2 := fileR.Chunks[i]
		if c1.Start != c2.Start || c1.Index != c2.Index || c1.End != c2.End {
			t.Error("Chunks don't match!")
		}
	}
}

func TestEmptyDirectory(t *testing.T) {
	file := File{
		URL: "http://test.fr/video.mp4",
	}
	file.setOutput("", false)

	if file.Output != "video.mp4" {
		t.Error("Directory should be equal to the filename")
	}
}

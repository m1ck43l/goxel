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
			{
				Start: 0,
				End:   99,
				Done:  0,
				Total: 100,
			},
			{
				Start: 100,
				End:   199,
				Done:  0,
				Total: 100,
			},
			{
				Start: 200,
				End:   299,
				Done:  0,
				Total: 100,
			},
		},
	}
	file.writeMetadata()

	fileR := File{
		Output:     path.Join(dir, "work.mp4"),
		OutputWork: path.Join(dir, "work.mp4."+workExtension),
	}
	fileR.ResumeChunks(3)

	sort.SliceStable(fileR.Chunks, func(i, j int) bool {
		return fileR.Chunks[i].Start < fileR.Chunks[j].Start
	})

	for i, c1 := range file.Chunks {
		c2 := fileR.Chunks[i]
		if c1.Start != c2.Start || c1.End != c2.End {
			t.Error("Chunks don't match!")
		}
	}
}

func TestResumeWithUpdateMaxConn(t *testing.T) {
	dir, err := ioutil.TempDir("", "goxel-test")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	file := File{
		Output:     path.Join(dir, "work.mp4"),
		OutputWork: path.Join(dir, "work.mp4."+workExtension),
		Chunks: []Chunk{
			{
				Start: 0,
				End:   99,
				Done:  50,
				Total: 100,
			},
			{
				Start: 100,
				End:   199,
				Done:  50,
				Total: 100,
			},
			{
				Start: 200,
				End:   299,
				Done:  0,
				Total: 100,
			},
		},
	}
	file.writeMetadata()

	fileR := File{
		Output:     path.Join(dir, "work.mp4"),
		OutputWork: path.Join(dir, "work.mp4."+workExtension),
	}
	fileR.ResumeChunks(4)

	sort.SliceStable(fileR.Chunks, func(i, j int) bool {
		return fileR.Chunks[i].Start < fileR.Chunks[j].Start
	})

	if len(fileR.Chunks) != 4 {
		t.Error("File should have 4 chunks after resume")
	}

	chkI := file.Chunks[2]
	chkR1 := fileR.Chunks[2]
	chkR2 := fileR.Chunks[3]

	if chkI.Start != chkR1.Start || chkI.End != chkR2.End || chkI.Total != chkR1.Total+chkR2.Total {
		t.Error("Fail to resume!")
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

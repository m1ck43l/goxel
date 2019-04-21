package main

import (
	"io/ioutil"
	"log"

	"github.com/m1ck43l/goxel/goxel"
)

func main() {
	log.SetOutput(ioutil.Discard)

	// Create a new GoXel instance and run it.
	goxel := goxel.NewGoXel()
	goxel.Run()
}

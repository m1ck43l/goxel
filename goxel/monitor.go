package goxel

import (
	"fmt"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

const monitorCount = 10

type monitor struct {
	Duration time.Duration
	Value    uint64
}

// Monitoring monitors the current downloads and display the speed and progress for each files
func Monitoring(files []*File, done chan bool) {
	monitors := make([]monitor, monitorCount, monitorCount)

	var count, pDone, gDone uint64
	lastStart := time.Now()

	move := 0
	for {
		select {
		default:
			gDone = 0

			if count > 0 {
				fmt.Printf(strings.Repeat("\033[F", move))
			}
			fmt.Printf("\r")

			if count == 0 {
				fmt.Printf("\n")
			}

			move = 0
			for idx, f := range files {
				if f.Error == "" {
					fmt.Printf("[%3d] - %-120v\n", idx, f.Output)
				} else {
					fmt.Printf("[ERR] - %v: %v\n", f.Output, f.Error)
				}
				move++
			}
			fmt.Printf("\n")
			move++

			var curDone uint64
			var curDelay time.Duration
			for _, vd := range monitors {
				curDone += vd.Value
				curDelay += vd.Duration
			}

			speed := uint64(float64(curDone) / (float64(curDelay/time.Nanosecond) / 1000000000))

			fmt.Printf("Download speed: %8v/s\n", humanize.Bytes(speed))
			fmt.Printf("\n")
			move += 2

			for idx, f := range files {
				if !f.Valid {
					continue
				}

				ratio, conn, done := f.UpdateStatus()

				fmt.Printf("[%3d] - [%6.2f%%] [%-101v] (%d/%d)\n", idx, ratio, strings.Repeat("=", int(ratio))+">", conn, len(f.Chunks))
				move++

				gDone += done
			}

			monitors[count%monitorCount] = monitor{
				Duration: time.Since(lastStart),
				Value:    gDone - pDone,
			}
			count++
			pDone = gDone
			lastStart = time.Now()

			time.Sleep(100 * time.Millisecond)

		case <-done:
			return
		}
	}
}

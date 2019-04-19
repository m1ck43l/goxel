package goxel

import (
	"fmt"
	"math"
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
func Monitoring(files []*File, done chan bool, quiet bool) {
	monitors := make([]monitor, monitorCount, monitorCount)

	var count, pDone, gDone uint64
	var output []string

	lastStart := time.Now()

	for {
		select {
		default:
			gDone = 0

			move := math.Max(float64(len(output)-1), 0)
			output = make([]string, 0)
			output = append(output, fmt.Sprintf(strings.Repeat("\033[F", int(move)))+"\r")

			for idx, f := range files {
				if f.Error == "" {
					output = append(output, fmt.Sprintf("[%3d] - %-120v", idx, f.Output))
				} else {
					output = append(output, fmt.Sprintf("[ERR] - %v: %v", f.Output, f.Error))
				}
			}
			output = append(output, "")

			var curDone uint64
			var curDelay time.Duration
			for _, vd := range monitors {
				curDone += vd.Value
				curDelay += vd.Duration
			}

			speed := uint64(float64(curDone) / (float64(curDelay/time.Nanosecond) / 1000000000))

			output = append(output, fmt.Sprintf("Download speed: %8v/s", humanize.Bytes(speed)))
			output = append(output, fmt.Sprintf("Active connections: %6v", activeConnections.v))
			output = append(output, "")

			for idx, f := range files {
				if !f.Valid {
					continue
				}

				ratio, conn, done := f.UpdateStatus()

				left := fmt.Sprintf("[%3d] - [%6.2f%%] [", idx, ratio)
				right := fmt.Sprintf("] (%d/%d)", conn, len(f.Chunks))

				unit := float64(int(getWidth())-len(left)-len(right)) / float64(f.Size)
				output = append(output, left+f.BuildProgress(unit)+right)

				gDone += done
			}
			output = append(output, "")

			monitors[count%monitorCount] = monitor{
				Duration: time.Since(lastStart),
				Value:    gDone - pDone,
			}
			count++
			pDone = gDone
			lastStart = time.Now()

			if !quiet {
				for _, s := range output {
					if s == "" {
						fmt.Printf("%v", strings.Repeat(" ", int(getWidth())))
					} else {
						fmt.Print(s + "\n")
					}
				}
			}

			time.Sleep(100 * time.Millisecond)

		case <-done:
			return
		}
	}
}

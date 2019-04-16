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
			output = append(output, "")

			for idx, f := range files {
				if !f.Valid {
					continue
				}

				ratio, conn, done := f.UpdateStatus()

				left := fmt.Sprintf("[%3d] - [%6.2f%%] [", idx, ratio)
				right := fmt.Sprintf("] (%d/%d)", conn, len(f.Chunks))

				c := float64(int(float64(int(getWidth())-len(left)-len(right)) / float64(len(f.Chunks))))

				progress := ""
				for i, chunk := range f.Chunks {
					offset := float64(len(fmt.Sprintf("%d", i)))

					var cInitial int
					if chunk.Initial > 0 {
						cInitial = int(math.Min(float64(chunk.Initial)/float64(chunk.Total)*c, c-offset))
					}

					var cRemaining int
					if chunk.Done < chunk.Total {
						cRemaining = int(math.Min(math.Max(float64(chunk.Total-chunk.Done)/float64(chunk.Total)*c, 0), c-offset))
					}

					cDone := int(math.Max(float64(int(c)-cInitial-cRemaining-int(offset)), 0))

					progress += fmt.Sprintf("%v%v%d%v", strings.Repeat("+", cInitial), strings.Repeat("-", cDone), i, strings.Repeat(" ", cRemaining))
				}

				output = append(output, left+progress+right)

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

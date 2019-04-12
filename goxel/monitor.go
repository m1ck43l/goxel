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

type status struct {
	Ratio            float64
	ActiveConnection uint64
	ChunksCount      int
}

func printMonitoring(move int, count uint64, speed uint64, files []*File, status []status) int {
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

	fmt.Printf("Download speed: %8v/s\n", humanize.Bytes(speed))
	fmt.Printf("\n")
	move += 2

	for idx, p := range status {
		fmt.Printf("[%3d] - [%6.2f%%] [%-101v] (%d/%d)\n", idx, p.Ratio, strings.Repeat("=", int(p.Ratio))+">", p.ActiveConnection, p.ChunksCount)
		move++
	}

	return move
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

			var curDone uint64
			var curDelay time.Duration
			for _, vd := range monitors {
				curDone += vd.Value
				curDelay += vd.Duration
			}

			speed := uint64(float64(curDone) / (float64(curDelay/time.Nanosecond) / 1000000000))

			s := make([]status, len(files), len(files))
			for _, f := range files {
				if !f.Valid {
					continue
				}

				var done, total, conn uint64
				for i := 0; i < len(f.Chunks); i++ {
					v := f.Chunks[i]
					done += v.Done
					total += v.Total

					if v.Done < v.Total && v.Done > v.Initial {
						conn++
					}
				}

				if done >= total && done > 0 && !f.Finished {
					f.finish()
				}

				if done > 0 && f.OutputWork != "" && !f.Finished {
					f.writeMetadata()
				}

				var ratio float64
				if total > 0 {
					ratio = float64(done) / float64(total) * 100
				}

				s[0] = status{
					Ratio:            ratio,
					ActiveConnection: conn,
					ChunksCount:      len(f.Chunks),
				}

				gDone += done
			}

			monitors[count%monitorCount] = monitor{
				Duration: time.Since(lastStart),
				Value:    gDone - pDone,
			}
			count++
			pDone = gDone
			lastStart = time.Now()

			move = printMonitoring(move, count, speed, files, s)

			time.Sleep(100 * time.Millisecond)

		case <-done:
			return
		}
	}
}

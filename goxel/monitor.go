package goxel

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

const monitorCount = 40

type monitor struct {
	Duration time.Duration
	Value    uint64
}

type monitorer interface {
	monitor(files []*File, d chan download, messages []string) (int, []string)
}

// QuietMonitoring only ensures the Files are synced every Xs
type QuietMonitoring struct {
	count uint64
}

func (q *QuietMonitoring) monitor(files []*File, d chan download, messages []string) (int, []string) {
	finished := 0
	for _, f := range files {
		if !f.Valid {
			if f.Error != "" {
				finished++
			}
			continue
		}

		f.UpdateStatus(q.count%10 == 0)
		if f.Finished {
			finished++
		}
	}
	q.count++

	return finished, make([]string, 0)
}

func buildFileDescription(output []string, files []*File) []string {
	for idx, f := range files {
		if f.Error == "" {
			output = append(output, fmt.Sprintf("[%3d] - %-120v", idx, f.Output))
		} else {
			output = append(output, fmt.Sprintf("[ERR] - %v: %v", f.Output, f.Error))
		}
	}
	return output
}

func buildChunkDescription(output []string, files []*File, count uint64, speed uint64) ([]string, int, uint64) {
	finished := 0
	var gDone uint64

	for idx, f := range files {
		if !f.Valid {
			if f.Error != "" {
				finished++
			}
			continue
		}

		ratio, conn, done, sdone := f.UpdateStatus(count%10 == 0)
		if f.Finished {
			finished++
		}

		left := fmt.Sprintf("[%3d] - [%6.2f%%] [", idx, ratio)

		var remaining uint64
		if speed > 0 {
			remaining = uint64(math.Max(float64(f.Size)-float64(done), 0)) / speed
		}
		right := fmt.Sprintf("] (%d/%d) [%8v]", conn, len(f.Chunks), fmtDuration(remaining))

		unit := float64(int(getWidth())-len(left)-len(right)-1) / float64(f.Size)
		output = append(output, left+f.BuildProgress(unit)+right)

		gDone += sdone
	}

	return output, finished, gDone
}

// ConsoleMonitoring monitors the current downloads and display the speed and progress for each files
type ConsoleMonitoring struct {
	monitors            []monitor
	count, pDone, gDone uint64
	output              []string
	lastStart           time.Time
}

func (c *ConsoleMonitoring) monitor(files []*File, d chan download, messages []string) (int, []string) {
	if goxel.Scroll {
		c.output = make([]string, 0)
	} else {
		move := math.Max(float64(len(c.output)-1), 0)
		c.output = make([]string, 0)
		c.output = append(c.output, fmt.Sprintf(strings.Repeat("\033[F", int(move)))+"\r")
	}

	for _, message := range messages {
		c.output = append(c.output, message)
	}
	c.output = append(c.output, "")

	c.output = buildFileDescription(c.output, files)
	c.output = append(c.output, "")

	var curDone uint64
	var curDelay time.Duration
	for _, vd := range c.monitors {
		curDone += vd.Value
		curDelay += vd.Duration
	}

	speed := uint64(float64(curDone) / (float64(curDelay/time.Nanosecond) / 1000000000))

	c.output = append(c.output, fmt.Sprintf("Download speed: %8v/s", humanize.Bytes(speed)))
	c.output = append(c.output, fmt.Sprintf("Active connections: %6v", activeConnections.v))
	c.output = append(c.output, "")

	var finished int
	c.output, finished, c.gDone = buildChunkDescription(c.output, files, c.count, speed)

	c.output = append(c.output, "")

	c.monitors[c.count%monitorCount] = monitor{
		Duration: time.Since(c.lastStart),
		Value:    c.gDone - c.pDone,
	}
	c.count++
	c.pDone = c.gDone
	c.lastStart = time.Now()

	for _, s := range c.output {
		if s == "" {
			fmt.Printf("%v", strings.Repeat(" ", int(getWidth())))
		} else {
			fmt.Print(s + "\n")
		}
	}

	return finished, messages
}

// Monitoring handles the files' termination and monitoring
func Monitoring(files []*File, done chan bool, d chan download, quiet bool) {
	var m monitorer
	if quiet {
		m = &QuietMonitoring{}
	} else {
		m = &ConsoleMonitoring{
			monitors:  make([]monitor, monitorCount),
			lastStart: time.Now(),
		}
	}

	gMessages := make([]string, 0)
	closed := false

	for {
		select {
		default:
			var finished int
			finished, gMessages = m.monitor(files, d, gMessages)
			if finished == len(files) && !closed {
				close(d)
				closed = true
			}
			time.Sleep(100 * time.Millisecond)

		case s := <-cMessages:
			if s.FileID == maxUint32 {
				gMessages = append(gMessages, fmt.Sprintf("[%v] - %7v - %v", s.Context, s.Type.String(), s.Content))
			} else {
				for _, file := range files {
					if file.ID == s.FileID {
						file.Error = s.Content
					}
				}
			}

		case <-done:
			return
		}
	}
}

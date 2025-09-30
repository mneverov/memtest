package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/trace"
	"strings"
	"time"
)

const (
	GB = 1_000_000_000
	MB = 1_000_000
)

var data [][]byte

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	fr := startFlightRecorder()
	run(fr)
}

func startFlightRecorder() *trace.FlightRecorder {
	// Set up the flight recorder
	fr := trace.NewFlightRecorder(trace.FlightRecorderConfig{
		MinAge:   10 * time.Second,
		MaxBytes: 10 << 20, // 10 MiB
	})
	fr.Start()
	return fr
}

func run(fr *trace.FlightRecorder) {
	// no difference with or without memory limit
	debug.SetMemoryLimit(GB)
	produceGarbage(fr)
}

// produceGarbage allocates some.
// Run the app and print RSS of a process in MB.
// "RSS: resident set size, the non-swapped physical memory that a task has used"
// ps -eo pid,cmd | grep memtest | awk 'NR==1 {print $1}' | xargs pmap -x | grep total | awk '{print "total: " int($3/1024) "MB; RSS: " int($4/1024) "MB; dirty: " int($5/1024) "MB"}'
func produceGarbage(fr *trace.FlightRecorder) {
	for i := 0; ; i++ {
		data = make([][]byte, 1024)
		for k := range 1024 {
			go func() {
				addRandomString(data, k)
				time.Sleep(10 * time.Millisecond)
			}()
		}

		if i%50 == 0 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			// todo(mneverov): test with debug.FreeOSMemory()?

			fmt.Printf(
				"%dGB total; heap: %fMB; released %dMB\n", m.TotalAlloc/GB, float64(m.HeapAlloc)/MB, m.HeapReleased/MB,
			)
		}
		if i == 1000 {
			// Capture a snapshot.
			if fr.Enabled() {
				go captureSnapshot(fr)
			}
		}
	}
}

func addRandomString(data [][]byte, i int) {
	rf := rand.Float64()
	s := strings.Repeat(fmt.Sprintf("%f", rf), 2048)
	data[i] = []byte(s)
}

// captureSnapshot captures a flight recorder snapshot. Should be called only once.
func captureSnapshot(fr *trace.FlightRecorder) {
	f, err := os.Create("snapshot.trace")
	if err != nil {
		log.Printf("opening snapshot file %s failed: %s", f.Name(), err)
		return
	}
	defer f.Close() // ignore error

	// WriteTo writes the flight recorder data to the provided io.Writer.
	_, err = fr.WriteTo(f)
	if err != nil {
		log.Printf("writing snapshot to file %s failed: %s", f.Name(), err)
		return
	}

	// Stop the flight recorder after the snapshot has been taken.
	fr.Stop()
	log.Printf("captured a flight recorder snapshot to %s", f.Name())
}

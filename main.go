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
	"strconv"
	"strings"
	"time"
	_ "unsafe"
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
	args := parseArgs()

	run(args, fr)
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

func run(args args, fr *trace.FlightRecorder) {
	// no difference with or without memory limit
	debug.SetMemoryLimit(GB)
	produceGarbage(args, fr)
}

// produceGarbage allocates some.
// Run the app and print RSS of a process in MB.
// "RSS: resident set size, the non-swapped physical memory that a task has used"
// ps -eo pid,cmd | grep memtest | awk 'NR==1 {print $1}' | xargs pmap -x | grep total | awk '{print "total: " int($3/1024) "MB; RSS: " int($4/1024) "MB; dirty: " int($5/1024) "MB"}'
func produceGarbage(args args, fr *trace.FlightRecorder) {
	stopCh := make(chan struct{})
	for i := 0; ; i++ {
		dataLen := 1024
		data = make([][]byte, dataLen)
		// Modify shared data concurrently. Should be ok since adding strings to different indexes.
		for idx := range dataLen {
			gCnt := runtime.NumGoroutine()
			if gCnt < 5 {
				log.Printf("%d goroutines\n", gCnt)
			}
			go func() {
				addRandomStringToIdx(data, idx)
				time.Sleep(10 * time.Millisecond)
			}()
		}

		if i%args.printIter == 0 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			// todo(mneverov): test with debug.FreeOSMemory()?
			log.Printf(
				"%dGB total; heap: %fMB; released %dMB\n", m.TotalAlloc/GB, float64(m.HeapAlloc)/MB, m.HeapReleased/MB,
			)
		}

		if shouldRecord(i, args.recordIter, args.autoRecord, fr.Enabled()) {
			go captureSnapshot(args.snapshotOut, fr, stopCh)
		}

		if shouldStop(i, args.stopIter, args.autoStop, stopCh) {
			log.Printf("stop at %d iteration\n", i)
			return
		}
	}
}

func shouldRecord(currIter, recordIter int, autoRecord, frEnabled bool) bool {
	if !frEnabled {
		return false
	}

	if currIter == recordIter && !autoRecord {
		return true
	}

	if autoRecord && runtime.NumGoroutine() == 1 {
		return true
	}

	return false
}

func shouldStop(currIter, stopIter int, autoStop bool, stopCh chan struct{}) bool {
	if currIter == stopIter {
		return true
	}

	if !autoStop {
		return false
	}

	select {
	case <-stopCh:
		return true
	default:
		return false
	}
}

func addRandomStringToIdx(data [][]byte, i int) {
	rf := rand.Float64()
	s := strings.Repeat(fmt.Sprintf("%f", rf), 2048)
	data[i] = []byte(s)
}

// captureSnapshot captures a flight recorder snapshot. Should be called only once.
func captureSnapshot(out string, fr *trace.FlightRecorder, ch chan struct{}) {
	start := time.Now()
	log.Printf("start capturing at %s\n", start.Format(time.RFC3339))
	var (
		err error
		f   *os.File
	)
	defer func() {
		if err == nil {
			log.Printf("captured a flight recorder snapshot in %v to %s", time.Now().Sub(start), out)
			close(ch)
		}
	}()

	f, err = os.Create(out)
	if err != nil {
		log.Printf("opening snapshot file %s failed: %s", out, err)
		return
	}
	defer f.Close() // ignore error

	_, err = fr.WriteTo(f)
	if err != nil {
		log.Printf("writing snapshot to file %s failed: %s", f.Name(), err)
		return
	}

	fr.Stop()
}

type args struct {
	stopIter    int
	recordIter  int
	printIter   int
	snapshotOut string
	autoRecord  bool
	autoStop    bool
}

var defaultArgs = args{
	stopIter:    -1,
	recordIter:  -1,
	printIter:   50,
	snapshotOut: "snapshot.trace",
	autoRecord:  false,
	autoStop:    false,
}

func parseArgs() args {
	res := defaultArgs

	for _, arg := range os.Args {
		if iter, ok := strings.CutPrefix(arg, "--stop-iter="); ok {
			res.stopIter = mustAtoi(iter)
		} else if iter, ok := strings.CutPrefix(arg, "--record-iter="); ok {
			res.recordIter = mustAtoi(iter)
		} else if iter, ok := strings.CutPrefix(arg, "--print-iter="); ok {
			res.printIter = mustAtoi(iter)
		} else if out, ok := strings.CutPrefix(arg, "--fr-out="); ok {
			res.snapshotOut = out
		} else if arg == "--auto-record" {
			res.autoRecord = true
		} else if arg == "--auto-stop" {
			res.autoStop = true
		}
	}

	log.Printf("stopIter=%d, recordIter=%d, printIter=%d, snapshotOut=%s, autoRecord=%t, autoStop=%t\n",
		res.stopIter, res.recordIter, res.printIter, res.snapshotOut, res.autoRecord, res.autoStop)

	return res
}

func mustAtoi(s string) int {
	res, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return res
}

package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"runtime/debug"
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
	memLimit()
}

func memLimit() {
	// no difference with or without memory limit
	debug.SetMemoryLimit(GB)
	garbage()
}

// garbage allocates some.
// Run the app and print RSS of a process in MB.
// "RSS: resident set size, the non-swapped physical memory that a task has used"
// ps -eo pid,cmd | grep memtest | awk 'NR==1 {print $1}' | xargs pmap -x | grep total | awk '{print "total: " int($3/1024) "MB; RSS: " int($4/1024) "MB; dirty: " int($5/1024) "MB"}'
func garbage() {
	for i := range 20000001 {
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
	}
}

func addRandomString(data [][]byte, i int) {
	rf := rand.Float64()
	s := strings.Repeat(fmt.Sprintf("%f", rf), 2048)
	data[i] = []byte(s)
}

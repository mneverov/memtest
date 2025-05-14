package main

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

var data []byte

func main() {
	memLimit()
}

const l00_MB = 100_000_000
const GB = 1_000_000_000

// garbage allocates some.
// go build -o mymem .
// ps -eo pid,cmd | grep mymem | awk 'NR==1 {print $1}' | xargs pmap | grep total
// GOGC=off mymem
func garbage() {
	for i := range 20000001 {
		data = make([]byte, l00_MB)
		data[0] = 1
		if i%20 == 0 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			//debug.FreeOSMemory()

			fmt.Printf("%dGB(total) - %dGB(released) = %dGB\n",
				m.TotalAlloc/GB, m.HeapReleased/GB, (m.TotalAlloc-m.HeapReleased)/GB)
		}
	}
}

func memLimit() {
	debug.SetMemoryLimit(GB)
	garbage()
}

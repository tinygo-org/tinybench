//go:build unix

package tinybench

import (
	"os"
	"runtime"
	"syscall"
)

// peakRSSKB returns the peak RSS of a finished process in KiB, see getrusage(2).
// On Linux it never reads below the harness RSS, which the child inherits at exec.
func peakRSSKB(ps *os.ProcessState) int64 {
	ru, ok := ps.SysUsage().(*syscall.Rusage)
	if !ok {
		return 0
	}
	if runtime.GOOS == "darwin" {
		return int64(ru.Maxrss) / 1024
	}
	return int64(ru.Maxrss)
}

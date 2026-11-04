//go:build !unix

package tinybench

import "os"

func peakRSSKB(ps *os.ProcessState) int64 { return 0 }

package main

import (
	_ "embed"
	"strings"
	"testing"
)

//go:embed test.txt
var input string

func TestParseBench(t *testing.T) {
	l, err := parsebench(strings.NewReader(input), "tinygo")
	if err != nil {
		t.Fatal(err)
	}
	if len(l) != 5 {
		t.Fatal("missing data")
	}
}

func TestParseBenchExtraMetrics(t *testing.T) {
	const in = `    bench_test.go:145: name="n-body" compiler="go" binarysize=2421227 version=1.26.2
BenchmarkAll/n-body:args=50000/go/go-32         	     100	   5991823 ns/op	 6012000 user-ns/op	   41000 sys-ns/op	    2048 peak-KiB
`
	b, err := parseBenchRaw(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	if len(b) != 1 || len(b[0].Runs) != 1 || b[0].Runs[0].PerOp != 5991823 {
		t.Fatalf("unexpected parse result %+v", b)
	}
}

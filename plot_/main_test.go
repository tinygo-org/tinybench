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

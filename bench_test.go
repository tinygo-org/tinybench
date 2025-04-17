package tinybench

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
)

type Compiler struct {
	Language     string
	CanRun       bool
	Compiler     string
	OutputBinary string
	MakeArgs     func(testname string) []string
}

var compilers = []Compiler{
	{
		Language:     "go",
		CanRun:       exec.Command("go", "version").Run() == nil,
		Compiler:     "go",
		OutputBinary: "./go.bin",
		MakeArgs: func(testname string) []string {
			return append(goBaseFlags, "./"+testname+"/go/main.go")
		},
	},
	{
		Language:     "go",
		CanRun:       exec.Command("tinygo", "version").Run() == nil,
		Compiler:     "tinygo",
		OutputBinary: "./tinybin",
		MakeArgs: func(testname string) []string {
			return append(tinygoBaseFlags, "./"+testname+"/go/main.go")
		},
	},
	{
		Language:     "c",
		CanRun:       exec.Command("gcc", "--version").Run() == nil,
		Compiler:     "gcc",
		OutputBinary: "./c.bin",
		MakeArgs:     cFlags,
	},
	{
		Language:     "c",
		CanRun:       exec.Command("clang", "--version").Run() == nil,
		Compiler:     "clang",
		OutputBinary: "./c.bin",
		MakeArgs:     cFlags,
	},
	{
		Language:     "zig",
		CanRun:       exec.Command("zig", "version").Run() == nil,
		Compiler:     "zig",
		OutputBinary: "./zig.bin",
		MakeArgs: func(testname string) []string {
			return append(zigBaseFlags, "./"+testname+"/zig/main.zig")
		},
	},
}

func cFlags(testname string) []string {
	linkFlags := gccLinkFlags[testname]
	cFilepath := "./" + testname + "/c/main.c"
	ccFlags := append(gccBaseFlags, cFilepath)
	return append(ccFlags, linkFlags...)
}

func BenchmarkAll(b *testing.B) {
	benchnames := setup()
	for _, c := range compilers {
		if c.CanRun {
			b.Logf("found compiler %q", c.Compiler)
		} else {
			b.Logf("skipping all benchmarks for compiler %q", c.Compiler)
		}
	}
	b.Logf("looking for benchmarks in %v", benchnames)
	for _, testname := range benchnames {
		argdata, err := os.ReadFile(testname + "/args.txt")
		casesJoined := strings.TrimSpace(string(argdata))
		if len(argdata) == 0 {
			b.Fatalf("%s has empty 'args.txt' file", testname)
		} else if err != nil {
			b.Fatalf("%s failed open arguments file 'args.txt': %s", testname, err)
		}

		cases := strings.Split(casesJoined, "\n")
		for _, compiler := range compilers {
			if !compiler.CanRun {
				continue // skip compiler benchmark.
			}
			testDir := testname + "/" + compiler.Language
			_, err := os.Stat(testDir)
			if os.IsNotExist(err) {
				b.Logf("%s skipped for %s", testname, compiler.Compiler)
				continue // Benchmark not implemented for this language.
			} else if err != nil {
				b.Fatal(err)
			}

			var onceCompile sync.Once
			ensureCompile := func(b *testing.B) {
				onceCompile.Do(func() {
					compArgs := compiler.MakeArgs(testname)
					out, err := exec.Command(compiler.Compiler, compArgs...).CombinedOutput()
					if err != nil {
						b.Fatalf("%s: building with %s flags=%v:\n%s", testname, compiler.Compiler, compArgs, out)
					}
					finfo, err := os.Stat(compiler.OutputBinary)
					if err != nil {
						b.Fatalf("%s: os.Stat(%q): %s", testname, compiler.OutputBinary, err.Error())
					}
					b.Logf("name=%q compiler=%q binarysize=%d\n", testname, compiler.Compiler, finfo.Size())
				})
			}
			for i := range cases {
				argInput := strings.Split(cases[i], " ")
				runBench(b, testname+":args="+cases[i]+"/"+compiler.Language+"/"+compiler.Compiler, compiler.OutputBinary, argInput, ensureCompile)
				if b.Failed() {
					b.FailNow()
				}
			}
		}
	}
}

func runBench(b *testing.B, name, binary string, benchFlags []string, ensureCompile func(b *testing.B)) {
	b.Helper()
	b.Run(name, func(b *testing.B) {
		var err error
		ensureCompile(b)
		for i := 0; i < b.N; i++ {
			err = exec.Command(binary, benchFlags...).Run()
			if err != nil {
				out, err2 := exec.Command(binary, benchFlags...).CombinedOutput()
				if err2 != nil {
					if err.Error() != err2.Error() {
						err2 = errors.Join(err, err2)
					}
					b.Fatalf("running program %s: %s\n%s", name, err2, out)
				}
				b.Fatalf("running program %s: %s", name, err)
			}
		}
	})
}

func setup() (benchnames []string) {
	fatal := func(msg string) {
		os.Stderr.WriteString(msg)
		os.Exit(1)
	}
	data, _ := os.ReadFile("go.mod")
	if !bytes.HasPrefix(data, []byte("module tinybench")) {
		fatal("run `go test` from root directory")

	}
	dirs, err := os.ReadDir(".")
	if err != nil {
		fatal(err.Error())
	}
	for _, dir := range dirs {
		if dir.IsDir() {
			name := dir.Name()
			if strings.ContainsAny(name, "._") {
				continue // skip
			}
			benchnames = append(benchnames, name)
		}
	}
	return benchnames
}

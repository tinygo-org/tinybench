package tinybench

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"testing"
)

type Compiler struct {
	Language       string
	VersionCommand *exec.Cmd
	Compiler       string
	OutputBinary   string
	MakeArgs       func(testname string) []string

	Version [3]int // 0:Major, 1:Minor, 2:Patch
}

func (c Compiler) CanRun() bool {
	return c.Version != [3]int{}
}

func (c Compiler) VersionString() string {
	return fmt.Sprintf("%d.%d.%d", c.Version[0], c.Version[1], c.Version[2])
}

var compilers = []Compiler{
	{
		Language:       "zig",
		VersionCommand: exec.Command("zig", "version"),
		Compiler:       "zig",
		OutputBinary:   "./zig.bin",
		MakeArgs: func(testname string) []string {
			return append(zigBaseFlags, "./"+testname+"/zig/main.zig")
		},
	},
	{
		Language:       "rust",
		VersionCommand: exec.Command("rustc", "-V"),
		Compiler:       "rustc",
		OutputBinary:   "./rust.bin",
		MakeArgs: func(testname string) []string {
			return append(rustBaseFlags, "./"+testname+"/rust/main.rs")
		},
	},
	{
		Language:       "go",
		VersionCommand: exec.Command("go", "version"),
		Compiler:       "go",
		OutputBinary:   "./go.bin",
		MakeArgs: func(testname string) []string {
			return append(goBaseFlags, "./"+testname+"/go/main.go")
		},
	},
	{
		Language:       "go",
		VersionCommand: exec.Command("tinygo", "version"),
		Compiler:       "tinygo",
		OutputBinary:   "./tinybin",
		MakeArgs: func(testname string) []string {
			return append(tinygoBaseFlags, "./"+testname+"/go/main.go")
		},
	},
	{
		Language:       "c",
		VersionCommand: exec.Command("gcc", "--version"),
		Compiler:       "gcc",
		OutputBinary:   "./c.bin",
		MakeArgs:       cFlags,
	},
	{
		Language:       "c",
		VersionCommand: exec.Command("clang", "--version"),
		Compiler:       "clang",
		OutputBinary:   "./c.bin",
		MakeArgs:       cFlags,
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
	for i, c := range compilers {
		version, err := c.VersionCommand.Output()
		if err != nil {
			b.Logf("skipping all benchmarks for compiler %q", c.Compiler)
			continue
		}
		var ok bool
		vMajor, vMinor, vPatch, ok := parseNextSemanticVersion(string(version))
		if !ok {
			b.Fatalf("unable to parse version for %s compiler from version output:\n%s", c.Compiler, version)
		}
		compilers[i].Version = [3]int{vMajor, vMinor, vPatch}
		b.Logf("found compiler %s %d.%d.%d", c.Compiler, vMajor, vMinor, vPatch)
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
			if !compiler.CanRun() {
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
					b.Logf("name=%q compiler=%q binarysize=%d version=%s\n", testname, compiler.Compiler, finfo.Size(), compiler.VersionString())
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
		b.StopTimer()
		ensureCompile(b)
		b.StartTimer()
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

// parseNExtSemanticVersion parses the first instance of anything that looks remotely close to a semantic version in the argument string
// and returns the major, minor and patch version as integers.
func parseNextSemanticVersion(s string) (int, int, int, bool) {
	firstDotIdx := strings.IndexByte(s, '.')
	if firstDotIdx <= 0 {
		return -1, -1, -1, false
	}
	i := firstDotIdx - 1
	for i >= 0 && isDigit(s[i]) {
		i--
	}
	start := i + 1
	vMajor, err := strconv.Atoi(s[start:firstDotIdx])
	if err != nil {
		return -1, -1, -1, false
	}
	i = firstDotIdx + 1
	for i < len(s) && isDigit(s[i]) {
		i++
	}
	secondDotIdx := i
	if s[secondDotIdx] != '.' {
		return -1, -1, -1, false
	}
	vMinor, err := strconv.Atoi(s[firstDotIdx+1 : secondDotIdx])
	if err != nil {
		return -1, -1, -1, false
	}
	i = secondDotIdx + 1
	for i < len(s) && isDigit(s[i]) {
		i++
	}
	vPatch, err := strconv.Atoi(s[secondDotIdx+1 : i])
	if err != nil {
		return -1, -1, -1, false
	}
	return vMajor, vMinor, vPatch, true
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

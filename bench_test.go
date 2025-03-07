package tinybench

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func BenchmarkAll(b *testing.B) {
	benchnames := setup()
	b.Logf("looking for benchmarks in %v", benchnames)
	hasClang := exec.Command("clang", "--version").Run() == nil
	hasZig := exec.Command("zig", "version").Run() == nil
	hasGo := exec.Command("go", "version").Run() == nil
	hasTinygo := exec.Command("tinygo", "version").Run() == nil
	for _, testname := range benchnames {
		argdata, err := os.ReadFile(testname + "/args.txt")
		casesJoined := strings.TrimSpace(string(argdata))
		if len(argdata) == 0 {
			b.Fatalf("%s has empty 'args.txt' file", testname)
		} else if err != nil {
			b.Fatalf("%s failed open arguments file 'args.txt': %s", testname, err)
		}
		cases := strings.Split(casesJoined, "\n")
		_, errGo := os.Stat(testname + "/go")
		_, errC := os.Stat(testname + "/c")
		_, errZig := os.Stat(testname + "/zig")
		for i := range cases {
			arginput := strings.Split(cases[i], " ")
			b.Run(testname+":args="+cases[i], func(b *testing.B) {
				// GO LANGUAGE BENCHMARKS.
				if hasGo && errGo == nil {
					// UPSTREAM GO BENCHMARK.
					goFilepath := "./" + testname + "/go/main.go"
					goFlags := append(goBaseFlags, goFilepath)
					runCompileAndBench(b, "go", "go", "./go.bin", goFlags, arginput)

					if hasTinygo {
						// TINYGO BENCHMARK.
						tinygoFlags := append(tinygoBaseFlags, goFilepath)
						runCompileAndBench(b, "tinygo", "tinygo", "./tinybin", tinygoFlags, arginput)
					}
				}

				// C LANGUAGE BENCHMARKS.
				if errC == nil {
					linkFlags := gccLinkFlags[testname]
					cFilepath := testname + "/c/main.c"
					gccFlags := append(gccBaseFlags, cFilepath)
					gccFlags = append(gccFlags, linkFlags...)
					// GCC COMPILER BENCHMARK.
					runCompileAndBench(b, "C gcc", "gcc", "./c.bin", gccFlags, arginput)

					if hasClang {
						// CLANG COMPILER BENCHMARK.
						runCompileAndBench(b, "C clang", "clang", "./c.bin", gccFlags, arginput)
					}
				}

				// ZIG LANGUAGE BENCHMARKS.
				if hasZig && errZig == nil {
					compilerFlags := append(zigBaseFlags, "./"+testname+"/zig/main.zig")
					runCompileAndBench(b, "zig", "zig", "./zig.bin", compilerFlags, arginput)
				}
			})
			if b.Failed() {
				b.FailNow() // Don't keep going if error encountered to avoid error spam on all benchmarks.
			}
		}
	}
}

func runCompileAndBench(b *testing.B, name, compiler, outputBinary string, compilerFlags, programFlags []string) {
	b.Helper()
	out, err := exec.Command(compiler, compilerFlags...).CombinedOutput()
	if err != nil {
		b.Fatalf("%s: building with %s flags=%v:\n%s", name, compilerFlags, compiler, out)
	}
	runBench(b, name, outputBinary, programFlags)
}

func runBench(b *testing.B, name, binary string, benchFlags []string) {
	b.Helper()
	b.Run(name, func(b *testing.B) {
		var err error
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

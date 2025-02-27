package tinybench

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func BenchmarkAll(b *testing.B) {
	benchnames := setup()
	b.Logf("looking for benchmarks in %v", benchnames)
	hasClang := exec.Command("clang", "--version").Run() == nil
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
		for i := range cases {
			arginput := strings.Split(cases[i], " ")
			b.Run(testname+":args="+cases[i], func(b *testing.B) {

				// GO PROGRAM.
				if errGo == nil {
					out, err := exec.Command("go", "build", "-o=go.bin", "./"+testname+"/go").CombinedOutput()
					if err != nil {
						b.Fatalf("building go: %s", out)
					}
					out, err = exec.Command("tinygo", "build", "-o=tinygo", "./"+testname+"/go").CombinedOutput()
					if err != nil {
						b.Fatalf("building tinygo: %s", out)
					}

					// GO COMPILER.
					b.Run("go", func(b *testing.B) {
						for i := 0; i < b.N; i++ {
							err = exec.Command("./go.bin", arginput...).Run()
							if err != nil {
								b.Fatalf("running go: %s", err)
							}
						}
					})

					// TINYGO COMPILER.
					b.Run("tinygo", func(b *testing.B) {
						for i := 0; i < b.N; i++ {
							err = exec.Command("./tinygo", arginput...).Run()
							if err != nil {
								b.Errorf("running tinygo: %s", err)
								b.Skip() // Maybe tinygo not compile?
							}
						}
					})
				}

				// C PROGRAM.
				if errC == nil {
					flags, ok := gccFlags[testname]
					if !ok {
						b.Fatalf("please add %s entry to gccFlags variable", testname)
					}
					args := strings.Split(flags, " ")
					out, err := exec.Command("gcc", args...).CombinedOutput()
					if err != nil {
						b.Fatalf("building with gcc: %s", out)
					}
					b.Run("C gcc", func(b *testing.B) {
						for i := 0; i < b.N; i++ {
							err = exec.Command("./c.bin", arginput...).Run()
							if err != nil {
								b.Errorf("running c: %s", err)
							}
						}
					})
					if hasClang {
						out, err := exec.Command("clang", args...).CombinedOutput()
						if err != nil {
							b.Fatalf("building with clang: %s", out)
						}
						b.Run("clang", func(b *testing.B) {
							for i := 0; i < b.N; i++ {
								err = exec.Command("./c.bin", arginput...).Run()
								if err != nil {
									b.Errorf("running c: %s", err)
								}
							}
						})
					}
				}
			})
		}
	}
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

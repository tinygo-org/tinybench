# tinybench
Benchmarks for comparing TinyGo's performance

## Run Benchmarks
```sh
go test -bench=.
```

#### Output for 12th Gen Intel(R) Core(TM) i5-12400F
```
$ go test -bench=. .
goos: linux
goarch: amd64
pkg: tinybench
cpu: 12th Gen Intel(R) Core(TM) i5-12400F
BenchmarkAll/rsa-keygen-s_1024/go-12                  92          14014770 ns/op
BenchmarkAll/rsa-keygen-s_1024/tinygo-12              69          38280000 ns/op
BenchmarkAll/rsa-keygen-s_1024/c-12                  135           8941168 ns/op
BenchmarkAll/rsa-keygen-s_2048/go-12                  14         106146057 ns/op
BenchmarkAll/rsa-keygen-s_2048/tinygo-12               4         740358439 ns/op
BenchmarkAll/rsa-keygen-s_2048/c-12                   31          44471414 ns/op
```

## Add a benchmark
The way tinybench works is all directories with no `.` or `_` character (anywhere in name) in this repos' root directory are added to the benchmark corpus.
Within each of these directories a `c` and `go` folder is searched for and their code compiled and run automatically. So adding a new benchmark is as simple as:

1. Creating a new top level folder with a descriptive name such as `rsa-keygen` with no `.` or `_` characters



3. Add an `args.txt` file to the folder with the OS arguments to the program and add a single line with an argument i.e: `-s 1024` (flag `s` with value `1024`).
    - Each line of this file will contain a test case

4. Create folders with the language you wish to test. Each will be run with arguments provided by `args.txt`
    - `go`: Will contain a `package main` project that is compiled.
    - `c`: Contains the C source code. Since linking is done via flags you must add your project's flags to `gccFlags` map in [`gccflags_test.go`]

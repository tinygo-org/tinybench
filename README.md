# tinybench
Benchmarks for comparing TinyGo's performance

## Benchmarks chosen and focus
- `fannkuch-redux`: Focused on integer operations on short arrays
- `n-body`: Floating point operations and usage of math library (sqrt)
- `rsa-keygen`: Usage of crypto library for secure key generation. C version uses OpenSSL while Go version uses standard library. Focus on speed of modern available crypto libraries.

## Run Benchmarks
```sh
go test -bench=.
```

#### Output for 12th Gen Intel(R) Core(TM) i5-12400F

```
$ go test -bench .
goos: linux
goarch: amd64
pkg: tinybench
cpu: 12th Gen Intel(R) Core(TM) i5-12400F
BenchmarkAll/fannkuch-redux:args=4/go-12            1368           2242250 ns/op
BenchmarkAll/fannkuch-redux:args=4/tinygo-12        1981            722743 ns/op
BenchmarkAll/fannkuch-redux:args=4/c-12             2438           1614646 ns/op
BenchmarkAll/fannkuch-redux:args=12/go-12              1        24182412496 ns/op
BenchmarkAll/fannkuch-redux:args=12/tinygo-12          1        23403250852 ns/op
BenchmarkAll/fannkuch-redux:args=12/c-12               1        29678815089 ns/op
BenchmarkAll/n-body:args=10000000/go-12                2         606381410 ns/op
BenchmarkAll/n-body:args=10000000/tinygo-12            3         420606198 ns/op
BenchmarkAll/n-body:args=10000000/c-12                 3         348006321 ns/op
BenchmarkAll/n-body:args=50000000/go-12                1        3025191449 ns/op
BenchmarkAll/n-body:args=50000000/tinygo-12            1        2106606611 ns/op
BenchmarkAll/n-body:args=50000000/c-12                 1        1742773380 ns/op
BenchmarkAll/rsa-keygen:args=-s_1024/go-12           100          12856298 ns/op
BenchmarkAll/rsa-keygen:args=-s_1024/tinygo-12                37          31143514 ns/op
BenchmarkAll/rsa-keygen:args=-s_1024/c-12                    111          10075440 ns/op
BenchmarkAll/rsa-keygen:args=-s_2048/go-12                    22          82791489 ns/op
BenchmarkAll/rsa-keygen:args=-s_2048/tinygo-12                 6         337273394 ns/op
BenchmarkAll/rsa-keygen:args=-s_2048/c-12                     62          34184541 ns/op
PASS
ok      tinybench       121.708s
```

## Result Summary
- TinyGo is notably faster at integer number crunching
- Both Go and TinyGo lag behind C at floating point math
- OpenSSL is notably faster than Go's standard library at RSA key generation

## Add a benchmark
The way tinybench works is all directories with no `.` or `_` character (anywhere in name) in this repos' root directory are added to the benchmark corpus.
Within each of these directories a `c` and `go` folder is searched for and their code compiled and run automatically. So adding a new benchmark is as simple as:

1. Creating a new top level folder with a descriptive name such as `rsa-keygen` with no `.` or `_` characters


2. Add an `args.txt` file to the folder with the OS arguments to the program and add a single line with an argument i.e: `-s 1024` (flag `s` with value `1024`).
    - Each line of this file will contain a test case

3. Create folders with the language you wish to test. Each will be run with arguments provided by `args.txt`
    - `go`: Will contain a `package main` project that is compiled.
    - `c`: Contains the C source code. Since linking is done via flags you must add your project's flags to `gccFlags` map in [`gccflags_test.go`]

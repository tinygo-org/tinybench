# tinybench
Benchmarks for comparing TinyGo's performance

## Benchmarks chosen and focus
- `fannkuch-redux`: Focused on integer operations on short arrays
- `n-body`: Floating point operations and usage of math library (sqrt)
    - `n-body-nosqrt`: Identical to above but replaces call to square-root math library function with a iterative solution. This benchmark shows the difference between C and Go math standard libraries. Go math library has more overhead for assembly implemented functions.
- `rsa-keygen`: Usage of crypto library for secure key generation. C version uses OpenSSL while Go version uses standard library. Focus on speed of modern available crypto libraries.

## Run Benchmarks
```sh
go test -bench=.

# Or to only run a certain test's benchmarks use expression "BenchmarkAll/<NAME OF TEST>:" 
go test -bench "BenchmarkAll/n-body:"  # You may need to escape the colon on windows powershell.
```

#### Output for 12th Gen Intel(R) Core(TM) i5-12400F

<details>
<summary>Click to display</summary>

```
$ go test -bench .
goos: linux
goarch: amd64
pkg: tinybench
cpu: 12th Gen Intel(R) Core(TM) i5-12400F
BenchmarkAll/fannkuch-redux:args=4/go-12            1502           1179924 ns/op
BenchmarkAll/fannkuch-redux:args=4/tinygo-12        5932            279511 ns/op
BenchmarkAll/fannkuch-redux:args=4/C_gcc-12         5388           1142870 ns/op
BenchmarkAll/fannkuch-redux:args=4/clang-12         4458           1126904 ns/op
BenchmarkAll/fannkuch-redux:args=8/go-12             198           6411570 ns/op
BenchmarkAll/fannkuch-redux:args=8/tinygo-12         303           3951872 ns/op
BenchmarkAll/fannkuch-redux:args=8/C_gcc-12          376           3402604 ns/op
BenchmarkAll/fannkuch-redux:args=8/clang-12          336           3443068 ns/op
BenchmarkAll/fannkuch-redux:args=10/go-12              6         168656329 ns/op
BenchmarkAll/fannkuch-redux:args=10/tinygo-12          7         148169104 ns/op
BenchmarkAll/fannkuch-redux:args=10/C_gcc-12           6         173121235 ns/op
BenchmarkAll/fannkuch-redux:args=10/clang-12           6         169094018 ns/op
BenchmarkAll/n-body:args=50000/go-12                 171           7944350 ns/op
BenchmarkAll/n-body:args=50000/tinygo-12             318           4422439 ns/op
BenchmarkAll/n-body:args=50000/C_gcc-12              264           4490991 ns/op
BenchmarkAll/n-body:args=50000/clang-12              351           4562434 ns/op
BenchmarkAll/n-body:args=100000/go-12                134           8982300 ns/op
BenchmarkAll/n-body:args=100000/tinygo-12            225           5591291 ns/op
BenchmarkAll/n-body:args=100000/C_gcc-12             238           5256402 ns/op
BenchmarkAll/n-body:args=100000/clang-12             231           5294096 ns/op
BenchmarkAll/n-body:args=1000000/go-12                16          67169729 ns/op
BenchmarkAll/n-body:args=1000000/tinygo-12            25          47041730 ns/op
BenchmarkAll/n-body:args=1000000/C_gcc-12             30          38944984 ns/op
BenchmarkAll/n-body:args=1000000/clang-12             30          39282334 ns/op
BenchmarkAll/n-body-nosqrt:args=50000/go-12           76          15167234 ns/op
BenchmarkAll/n-body-nosqrt:args=50000/tinygo-12                       93          12735899 ns/op
BenchmarkAll/n-body-nosqrt:args=50000/C_gcc-12                        93          12381674 ns/op
BenchmarkAll/n-body-nosqrt:args=50000/clang-12                        98          12339877 ns/op
BenchmarkAll/n-body-nosqrt:args=100000/go-12                          40          29178278 ns/op
BenchmarkAll/n-body-nosqrt:args=100000/tinygo-12                      48          24773376 ns/op
BenchmarkAll/n-body-nosqrt:args=100000/C_gcc-12                       46          23837935 ns/op
BenchmarkAll/n-body-nosqrt:args=100000/clang-12                       48          23809311 ns/op
BenchmarkAll/n-body-nosqrt:args=1000000/go-12                          4         286624158 ns/op
BenchmarkAll/n-body-nosqrt:args=1000000/tinygo-12                      5         247300422 ns/op
BenchmarkAll/n-body-nosqrt:args=1000000/C_gcc-12                       5         235582567 ns/op
BenchmarkAll/n-body-nosqrt:args=1000000/clang-12                       5         234883130 ns/op
BenchmarkAll/rsa-keygen:args=-s_512/go-12                            171           5871000 ns/op
BenchmarkAll/rsa-keygen:args=-s_512/tinygo-12                        267           5237971 ns/op
BenchmarkAll/rsa-keygen:args=-s_512/C_gcc-12                         250           4223293 ns/op
BenchmarkAll/rsa-keygen:args=-s_512/clang-12                         280           4145076 ns/op
BenchmarkAll/rsa-keygen:args=-s_1024/go-12                            86          14122028 ns/op
BenchmarkAll/rsa-keygen:args=-s_1024/tinygo-12                        46          29042923 ns/op
BenchmarkAll/rsa-keygen:args=-s_1024/C_gcc-12                        129           8804081 ns/op
BenchmarkAll/rsa-keygen:args=-s_1024/clang-12                        135           8640904 ns/op
BenchmarkAll/rsa-keygen:args=-s_2048/go-12                            12          86912016 ns/op
BenchmarkAll/rsa-keygen:args=-s_2048/tinygo-12                         3         579394173 ns/op
BenchmarkAll/rsa-keygen:args=-s_2048/C_gcc-12                         26          39559443 ns/op
BenchmarkAll/rsa-keygen:args=-s_2048/clang-12                         28          39168093 ns/op
```

</details>

## Result Summary
- TinyGo is notably faster at integer number crunching.
- TinyGo and C go head-to-head on floating point math when not calling specialized functions such as `sqrt`. Go lags behind.
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

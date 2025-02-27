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
BenchmarkAll/fannkuch-redux:args=4/go-12            1689           1761810 ns/op
BenchmarkAll/fannkuch-redux:args=4/tinygo-12        4015            462891 ns/op
BenchmarkAll/fannkuch-redux:args=4/C_gcc-12         4515            926869 ns/op
BenchmarkAll/fannkuch-redux:args=4/clang-12         3184           1408976 ns/op
BenchmarkAll/fannkuch-redux:args=8/go-12             319           4811605 ns/op
BenchmarkAll/fannkuch-redux:args=8/tinygo-12         368           3644272 ns/op
BenchmarkAll/fannkuch-redux:args=8/C_gcc-12          303           3598200 ns/op
BenchmarkAll/fannkuch-redux:args=8/clang-12          390           2722773 ns/op
BenchmarkAll/fannkuch-redux:args=10/go-12              6         167180506 ns/op
BenchmarkAll/fannkuch-redux:args=10/tinygo-12          7         163017421 ns/op
BenchmarkAll/fannkuch-redux:args=10/C_gcc-12           5         203131712 ns/op
BenchmarkAll/fannkuch-redux:args=10/clang-12           6         170950045 ns/op
BenchmarkAll/n-body:args=50000/go-12                 175           6997592 ns/op
BenchmarkAll/n-body:args=50000/tinygo-12             339           4332278 ns/op
BenchmarkAll/n-body:args=50000/C_gcc-12              382           3491287 ns/op
BenchmarkAll/n-body:args=50000/clang-12              357           3373237 ns/op
BenchmarkAll/n-body:args=100000/go-12                133           9573734 ns/op
BenchmarkAll/n-body:args=100000/tinygo-12            200           5378470 ns/op
BenchmarkAll/n-body:args=100000/C_gcc-12             228           5322187 ns/op
BenchmarkAll/n-body:args=100000/clang-12             234           4981824 ns/op
BenchmarkAll/n-body:args=1000000/go-12                16          67614088 ns/op
BenchmarkAll/n-body:args=1000000/tinygo-12            25          47178324 ns/op
BenchmarkAll/n-body:args=1000000/C_gcc-12             30          39153560 ns/op
BenchmarkAll/n-body:args=1000000/clang-12             30          39798239 ns/op
BenchmarkAll/n-body-nosqrt:args=50000/go-12           73          15274763 ns/op
BenchmarkAll/n-body-nosqrt:args=50000/tinygo-12                       81          12735334 ns/op
BenchmarkAll/n-body-nosqrt:args=50000/C_gcc-12                        97          12367592 ns/op
BenchmarkAll/n-body-nosqrt:args=50000/clang-12                        88          12445516 ns/op
BenchmarkAll/n-body-nosqrt:args=100000/go-12                          40          29376413 ns/op
BenchmarkAll/n-body-nosqrt:args=100000/tinygo-12                      45          24916553 ns/op
BenchmarkAll/n-body-nosqrt:args=100000/C_gcc-12                       50          24428093 ns/op
BenchmarkAll/n-body-nosqrt:args=100000/clang-12                       49          24357044 ns/op
BenchmarkAll/n-body-nosqrt:args=1000000/go-12                          4         288958087 ns/op
BenchmarkAll/n-body-nosqrt:args=1000000/tinygo-12                      5         245972148 ns/op
BenchmarkAll/n-body-nosqrt:args=1000000/C_gcc-12                       5         235981302 ns/op
BenchmarkAll/n-body-nosqrt:args=1000000/clang-12                       5         235178796 ns/op
BenchmarkAll/rsa-keygen:args=-s_512/go-12                            172           6350657 ns/op
BenchmarkAll/rsa-keygen:args=-s_512/tinygo-12                        205           5466632 ns/op
BenchmarkAll/rsa-keygen:args=-s_512/C_gcc-12                         252           4068652 ns/op
BenchmarkAll/rsa-keygen:args=-s_512/clang-12                         290           4103585 ns/op
BenchmarkAll/rsa-keygen:args=-s_1024/go-12                            88          13716162 ns/op
BenchmarkAll/rsa-keygen:args=-s_1024/tinygo-12                        30          38450349 ns/op
BenchmarkAll/rsa-keygen:args=-s_1024/C_gcc-12                        127           8914093 ns/op
BenchmarkAll/rsa-keygen:args=-s_1024/clang-12                        136           9019343 ns/op
BenchmarkAll/rsa-keygen:args=-s_2048/go-12                            14         100424873 ns/op
BenchmarkAll/rsa-keygen:args=-s_2048/tinygo-12                         2         640627691 ns/op
BenchmarkAll/rsa-keygen:args=-s_2048/C_gcc-12                         40          34515496 ns/op
BenchmarkAll/rsa-keygen:args=-s_2048/clang-12                         55          38689188 ns/op
PASS
ok      tinybench       112.482s
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

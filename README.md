# tinybench
Benchmarks for comparing TinyGo's performance

## Benchmarks chosen and focus
- `fannkuch-redux`: Focused on integer operations on short arrays
- `n-body`: Floating point operations and usage of math library (sqrt)
    - `n-body-nosqrt`: Identical to above but replaces call to square-root math library function with a iterative solution. This benchmark shows the difference between C and Go math standard libraries. Go math library has more overhead for assembly implemented functions.
- `rsa-keygen`: Usage of crypto library for secure key generation. C version uses OpenSSL while Go version uses standard library. Focus on speed of modern available crypto libraries.

![Benchmarks](./benchmark.png)

## Run Benchmarks
```sh
go test -bench=.

# Or to only run a certain test's benchmarks use expression "BenchmarkAll/<NAME OF TEST>:" 
go test -bench "BenchmarkAll/n-body:"  # You may need to escape the colon on windows powershell.
```

#### Generate benchmark image
Note the below command will not output results
```sh
go test -bench . | go run ./plot_/ -o benchmark.png
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
BenchmarkAll/fannkuch-redux:args=6/go-12            1351           1270129 ns/op
BenchmarkAll/fannkuch-redux:args=6/tinygo-12        7957            612381 ns/op
BenchmarkAll/fannkuch-redux:args=6/C_gcc-12         1310           1496906 ns/op
BenchmarkAll/fannkuch-redux:args=6/C_clang-12       2767           1199081 ns/op
BenchmarkAll/fannkuch-redux:args=7/go-12            1291           2350499 ns/op
BenchmarkAll/fannkuch-redux:args=7/tinygo-12        2554            948141 ns/op
BenchmarkAll/fannkuch-redux:args=7/C_gcc-12         1070           2049712 ns/op
BenchmarkAll/fannkuch-redux:args=7/C_clang-12       1298           2050549 ns/op
BenchmarkAll/fannkuch-redux:args=9/go-12              72          17253418 ns/op
BenchmarkAll/fannkuch-redux:args=9/tinygo-12          86          15095211 ns/op
BenchmarkAll/fannkuch-redux:args=9/C_gcc-12           56          19360511 ns/op
BenchmarkAll/fannkuch-redux:args=9/C_clang-12         63          16980935 ns/op
BenchmarkAll/n-body:args=50000/go-12                 168           7220572 ns/op
BenchmarkAll/n-body:args=50000/tinygo-12             283           5919548 ns/op
BenchmarkAll/n-body:args=50000/C_gcc-12              250           4544150 ns/op
BenchmarkAll/n-body:args=50000/C_clang-12            247           4815066 ns/op
BenchmarkAll/n-body:args=100000/go-12                100          10387685 ns/op
BenchmarkAll/n-body:args=100000/tinygo-12            164           8660934 ns/op
BenchmarkAll/n-body:args=100000/C_gcc-12             158           8040684 ns/op
BenchmarkAll/n-body:args=100000/C_clang-12           192           6362288 ns/op
BenchmarkAll/n-body:args=200000/go-12                 72          16455138 ns/op
BenchmarkAll/n-body:args=200000/tinygo-12            100          12196919 ns/op
BenchmarkAll/n-body:args=200000/C_gcc-12             100          10974046 ns/op
BenchmarkAll/n-body:args=200000/C_clang-12           100          10843887 ns/op
BenchmarkAll/n-body-nosqrt:args=50000/go-12           72          17247074 ns/op
BenchmarkAll/n-body-nosqrt:args=50000/tinygo-12                       88          14495706 ns/op
BenchmarkAll/n-body-nosqrt:args=50000/C_gcc-12                        91          14400444 ns/op
BenchmarkAll/n-body-nosqrt:args=50000/C_clang-12                      96          14531815 ns/op
BenchmarkAll/n-body-nosqrt:args=100000/go-12                          37          31447107 ns/op
BenchmarkAll/n-body-nosqrt:args=100000/tinygo-12                      44          26331575 ns/op
BenchmarkAll/n-body-nosqrt:args=100000/C_gcc-12                       46          26425921 ns/op
BenchmarkAll/n-body-nosqrt:args=100000/C_clang-12                     44          26123678 ns/op
BenchmarkAll/n-body-nosqrt:args=200000/go-12                          19          60256239 ns/op
BenchmarkAll/n-body-nosqrt:args=200000/tinygo-12                      21          51722668 ns/op
BenchmarkAll/n-body-nosqrt:args=200000/C_gcc-12                       21          49408678 ns/op
BenchmarkAll/n-body-nosqrt:args=200000/C_clang-12                     22          49472312 ns/op
BenchmarkAll/rsa-keygen:args=-s_512/go-12                            207           5776060 ns/op
BenchmarkAll/rsa-keygen:args=-s_512/tinygo-12                        224           5334605 ns/op
BenchmarkAll/rsa-keygen:args=-s_512/C_gcc-12                         232           5785600 ns/op
BenchmarkAll/rsa-keygen:args=-s_512/C_clang-12                       247           5078196 ns/op
BenchmarkAll/rsa-keygen:args=-s_1024/go-12                           100          14654308 ns/op
BenchmarkAll/rsa-keygen:args=-s_1024/tinygo-12                        31          34826760 ns/op
BenchmarkAll/rsa-keygen:args=-s_1024/C_gcc-12                        100          10635418 ns/op
BenchmarkAll/rsa-keygen:args=-s_1024/C_clang-12                      100          11462040 ns/op
PASS
ok      tinybench       104.670s
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

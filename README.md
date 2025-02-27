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
BenchmarkAll/fannkuch-redux:args=4/go-12            1542           1431544 ns/op
BenchmarkAll/fannkuch-redux:args=4/tinygo-12        2785            371839 ns/op
BenchmarkAll/fannkuch-redux:args=4/C_gcc-12         4214            851355 ns/op
BenchmarkAll/fannkuch-redux:args=4/clang-12         4060           1066186 ns/op
BenchmarkAll/fannkuch-redux:args=8/go-12             363           5324803 ns/op
BenchmarkAll/fannkuch-redux:args=8/tinygo-12         388           2784258 ns/op
BenchmarkAll/fannkuch-redux:args=8/C_gcc-12          518           3352672 ns/op
BenchmarkAll/fannkuch-redux:args=8/clang-12          346           3056667 ns/op
BenchmarkAll/fannkuch-redux:args=10/go-12              6         167958101 ns/op
BenchmarkAll/fannkuch-redux:args=10/tinygo-12          7         149447257 ns/op
BenchmarkAll/fannkuch-redux:args=10/C_gcc-12           5         204580862 ns/op
BenchmarkAll/fannkuch-redux:args=10/clang-12           6         172270561 ns/op
BenchmarkAll/n-body:args=50000/go-12                 166           6807364 ns/op
BenchmarkAll/n-body:args=50000/tinygo-12             290           3537776 ns/op
BenchmarkAll/n-body:args=50000/C_gcc-12              280           3891612 ns/op
BenchmarkAll/n-body:args=50000/clang-12              385           2757321 ns/op
BenchmarkAll/n-body:args=100000/go-12                138           9208135 ns/op
BenchmarkAll/n-body:args=100000/tinygo-12            200           5348460 ns/op
BenchmarkAll/n-body:args=100000/C_gcc-12             222           5472387 ns/op
BenchmarkAll/n-body:args=100000/clang-12             254           5624087 ns/op
BenchmarkAll/n-body:args=1000000/go-12                16          67452552 ns/op
BenchmarkAll/n-body:args=1000000/tinygo-12            25          46872229 ns/op
BenchmarkAll/n-body:args=1000000/C_gcc-12             30          38848753 ns/op
BenchmarkAll/n-body:args=1000000/clang-12             28          39366134 ns/op
BenchmarkAll/n-body-nosqrt:args=50000/go-12           79          15106833 ns/op
BenchmarkAll/n-body-nosqrt:args=50000/tinygo-12                       93          12727447 ns/op
BenchmarkAll/n-body-nosqrt:args=50000/C_gcc-12                        98          12286147 ns/op
BenchmarkAll/n-body-nosqrt:args=50000/clang-12                        98          12581121 ns/op
BenchmarkAll/n-body-nosqrt:args=100000/go-12                          40          29137500 ns/op
BenchmarkAll/n-body-nosqrt:args=100000/tinygo-12                      45          24970490 ns/op
BenchmarkAll/n-body-nosqrt:args=100000/C_gcc-12                       49          23967553 ns/op
BenchmarkAll/n-body-nosqrt:args=100000/clang-12                       48          23924742 ns/op
BenchmarkAll/n-body-nosqrt:args=1000000/go-12                          4         286036742 ns/op
BenchmarkAll/n-body-nosqrt:args=1000000/tinygo-12                      5         245628308 ns/op
BenchmarkAll/n-body-nosqrt:args=1000000/C_gcc-12                       5         234740616 ns/op
BenchmarkAll/n-body-nosqrt:args=1000000/clang-12                       5         235310196 ns/op
BenchmarkAll/rsa-keygen:args=-s_512/go-12                            189           5540392 ns/op
BenchmarkAll/rsa-keygen:args=-s_512/tinygo-12                        238           4851737 ns/op
BenchmarkAll/rsa-keygen:args=-s_512/C_gcc-12                         292           4053931 ns/op
BenchmarkAll/rsa-keygen:args=-s_512/clang-12                         285           4083646 ns/op
BenchmarkAll/rsa-keygen:args=-s_1024/go-12                            91          12839716 ns/op
BenchmarkAll/rsa-keygen:args=-s_1024/tinygo-12                        42          27442518 ns/op
BenchmarkAll/rsa-keygen:args=-s_1024/C_gcc-12                        128           9356074 ns/op
BenchmarkAll/rsa-keygen:args=-s_1024/clang-12                        126           8979752 ns/op
BenchmarkAll/rsa-keygen:args=-s_2048/go-12                            10         112742735 ns/op
BenchmarkAll/rsa-keygen:args=-s_2048/tinygo-12                         6         361506441 ns/op
BenchmarkAll/rsa-keygen:args=-s_2048/C_gcc-12                         26          51795005 ns/op
BenchmarkAll/rsa-keygen:args=-s_2048/clang-12                         32          39741682 ns/op
PASS
ok      tinybench       118.066s
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

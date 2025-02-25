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
BenchmarkAll/fannkuch-redux:args=4/go-12            2934           1256050 ns/op
BenchmarkAll/fannkuch-redux:args=4/tinygo-12        4084            479744 ns/op
BenchmarkAll/fannkuch-redux:args=4/c-12             5589           1409489 ns/op
BenchmarkAll/fannkuch-redux:args=8/go-12             190           7512802 ns/op
BenchmarkAll/fannkuch-redux:args=8/tinygo-12         264           6939621 ns/op
BenchmarkAll/fannkuch-redux:args=8/c-12              246           6902765 ns/op
BenchmarkAll/fannkuch-redux:args=10/go-12              7         154386257 ns/op
BenchmarkAll/fannkuch-redux:args=10/tinygo-12          7         154125962 ns/op
BenchmarkAll/fannkuch-redux:args=10/c-12               6         194099105 ns/op
BenchmarkAll/n-body:args=50000/go-12                 154           8137065 ns/op
BenchmarkAll/n-body:args=50000/tinygo-12             228           8293582 ns/op
BenchmarkAll/n-body:args=50000/c-12                  412           3389684 ns/op
BenchmarkAll/n-body:args=100000/go-12                134           9535106 ns/op
BenchmarkAll/n-body:args=100000/tinygo-12            181           7984095 ns/op
BenchmarkAll/n-body:args=100000/c-12                 162           6381976 ns/op
BenchmarkAll/n-body:args=1000000/go-12                18          63545177 ns/op
BenchmarkAll/n-body:args=1000000/tinygo-12            26          45244885 ns/op
BenchmarkAll/n-body:args=1000000/c-12                 30          38117632 ns/op
BenchmarkAll/n-body-nosqrt:args=50000/go-12           79          14999125 ns/op
BenchmarkAll/n-body-nosqrt:args=50000/tinygo-12                       93          12611197 ns/op
BenchmarkAll/n-body-nosqrt:args=50000/c-12                           100          12458944 ns/op
BenchmarkAll/n-body-nosqrt:args=100000/go-12                          39          29244188 ns/op
BenchmarkAll/n-body-nosqrt:args=100000/tinygo-12                      48          24458389 ns/op
BenchmarkAll/n-body-nosqrt:args=100000/c-12                           49          23534435 ns/op
BenchmarkAll/n-body-nosqrt:args=1000000/go-12                          4         287893191 ns/op
BenchmarkAll/n-body-nosqrt:args=1000000/tinygo-12                      5         243552483 ns/op
BenchmarkAll/n-body-nosqrt:args=1000000/c-12                           5         232477478 ns/op
BenchmarkAll/rsa-keygen:args=-s_512/go-12                            237           5636486 ns/op
BenchmarkAll/rsa-keygen:args=-s_512/tinygo-12                        207           5480214 ns/op
BenchmarkAll/rsa-keygen:args=-s_512/c-12                             297           4836271 ns/op
BenchmarkAll/rsa-keygen:args=-s_1024/go-12                           100          14437702 ns/op
BenchmarkAll/rsa-keygen:args=-s_1024/tinygo-12                        28          36936543 ns/op
BenchmarkAll/rsa-keygen:args=-s_1024/c-12                            128           9146223 ns/op
BenchmarkAll/rsa-keygen:args=-s_2048/go-12                            18          93172762 ns/op
BenchmarkAll/rsa-keygen:args=-s_2048/tinygo-12                         3         694935808 ns/op
BenchmarkAll/rsa-keygen:args=-s_2048/c-12                             31          37147406 ns/op
PASS
ok      tinybench       96.139s
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

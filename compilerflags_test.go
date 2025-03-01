package tinybench

var gccFlags = map[string]string{
	"rsa-keygen":     "-O3 -o c.bin rsa-keygen/c/main.c -lssl -lcrypto",
	"fannkuch-redux": "-O3 -o c.bin fannkuch-redux/c/main.c",
	"n-body":         "-O3 -o c.bin n-body/c/main.c",
	"n-body-nosqrt":  "-O3 -o c.bin n-body-nosqrt/c/main.c",
}

var zigBaseFlags = []string{
	"build-exe",
	"-femit-bin=zig.bin",
	"-fno-incremental",
	// Prominent Zig projects use ReleaseSafe instead of ReleaseFast.
	// This would seem to be a more realistic measure of how zig would perform
	// and also puts the compiler to the test of how well it can eliminate bounds checks.
	// https://github.com/tigerbeetle/tigerbeetle/blob/ae7f25dbd904f27498673bf2d60a51f21759cdb8/build.zig#L470
	"-O", "ReleaseSafe",
}

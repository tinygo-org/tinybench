package tinybench

var gccLinkFlags = map[string][]string{
	"n-body":        {"-lm"}, // Math library.
	"n-body-nosqrt": {"-lm"}, // Math library.
	"spectral-norm": {"-lm"}, // Math library.
}

var gccBaseFlags = []string{
	// "-O2"=="-O=ReleaseSafe" Same reasoning as with Zig.
	// It seems serious projects compile with memory limits and safety on.
	"-O2",
	"-o", "c.bin",
}

var zigBaseFlags = []string{
	"build-exe",
	"-femit-bin=zig.bin",
	"-fno-incremental",
	// Prominent Zig projects use ReleaseSafe instead of ReleaseFast.
	// This would seem to be a more realistic measure of how zig would perform in real circumstances.
	// Also puts the compiler to the test of how well it can eliminate bounds checks.
	// https://github.com/tigerbeetle/tigerbeetle/blob/ae7f25dbd904f27498673bf2d60a51f21759cdb8/build.zig#L470
	"-O", "ReleaseSafe",
}

var goBaseFlags = []string{
	"build",
	"-o=go.bin",
}

var tinygoBaseFlags = []string{
	"build",
	"-opt=2",
	"-o=tinybin",
}

var rustBaseFlags = []string{
	"-Copt-level=3",
	"-o", "rust.bin",
}

package tinybench

var gccFlags = map[string]string{
	"rsa-keygen":     "-O3 -o c.bin rsa-keygen/c/main.c -lssl -lcrypto",
	"fannkuch-redux": "-O3 -o c.bin fannkuch-redux/c/main.c",
	"n-body":         "-O3 -o c.bin n-body/c/main.c",
	"n-body-nosqrt":  "-O3 -o c.bin n-body-nosqrt/c/main.c",
}

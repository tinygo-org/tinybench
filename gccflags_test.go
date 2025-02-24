package tinybench

var gccFlags = map[string]string{
	"rsa-keygen": "-o c.bin rsa-keygen/c/main.c -lssl -lcrypto",
}

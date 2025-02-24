package main

import (
	"crypto/rand"
	"crypto/rsa"
	"flag"
	"os"
)

func main() {
	var size int
	flag.IntVar(&size, "s", 512, "Size of RSA key in bits")
	flag.Parse()
	_, err := rsa.GenerateKey(rand.Reader, size)
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}
}

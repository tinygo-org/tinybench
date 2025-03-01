package main

import (
	"fmt"
	"os"
	"strconv"
)

type pfannkuch struct {
	s, t     [16]elem
	maxflips int
	max_n    int
	odd      int
	checksum int
}

type elem int

func (pf *pfannkuch) flip() int {
	i := 1
	copy(pf.t[:], pf.s[:pf.max_n])
	for {
		// Reverse elements from t[0] to t[t[0]]
		x, y := 0, int(pf.t[0])
		for x < y {
			pf.t[x], pf.t[y] = pf.t[y], pf.t[x]
			x++
			y--
		}
		i++
		if pf.t[pf.t[0]] == 0 {
			break
		}
	}
	return i
}

func (pf *pfannkuch) rotate(n int) {
	c := pf.s[0]
	for i := 1; i <= n; i++ {
		pf.s[i-1] = pf.s[i]
	}
	pf.s[n] = c
}

func (pf *pfannkuch) tk(n int) {
	i := 0
	c := [16]elem{}
	for i < n {
		pf.rotate(i)
		if c[i] >= elem(i) {
			c[i] = 0
			i++
			continue
		}
		c[i]++
		i = 1
		pf.odd = ^pf.odd
		if pf.s[0] != 0 {
			f := 1
			if pf.s[pf.s[0]] != 0 {
				f = pf.flip()
			}
			if f > pf.maxflips {
				pf.maxflips = f
			}
			if pf.odd != 0 {
				pf.checksum -= f
			} else {
				pf.checksum += f
			}
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s number\n", os.Args[0])
		os.Exit(1)
	}
	var pf pfannkuch
	pf.max_n, _ = strconv.Atoi(os.Args[1])
	if pf.max_n < 3 || pf.max_n > 15 {
		fmt.Fprintf(os.Stderr, "max N range: must be 3 <= n <= 12\n")
		os.Exit(1)
	}

	for i := 0; i < pf.max_n; i++ {
		pf.s[i] = elem(i)
	}
	pf.tk(pf.max_n)

	fmt.Printf("%d\nPfannkuchen(%d) = %d\n", pf.checksum, pf.max_n, pf.maxflips)
}

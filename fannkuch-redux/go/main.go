package main

import (
	"fmt"
	"os"
	"strconv"
)

type elem int

var (
	s        [16]elem
	t        [16]elem
	maxflips int
	max_n    int
	odd      int
	checksum int
)

func flip() int {
	i := 1
	copy(t[:], s[:max_n])
	for {
		// Reverse elements from t[0] to t[t[0]]
		x, y := 0, int(t[0])
		for x < y {
			t[x], t[y] = t[y], t[x]
			x++
			y--
		}
		i++
		if t[t[0]] == 0 {
			break
		}
	}
	return i
}

func rotate(n int) {
	c := s[0]
	for i := 1; i <= n; i++ {
		s[i-1] = s[i]
	}
	s[n] = c
}

func tk(n int) {
	i := 0
	c := [16]elem{}
	for i < n {
		rotate(i)
		if c[i] >= elem(i) {
			c[i] = 0
			i++
			continue
		}
		c[i]++
		i = 1
		odd = ^odd
		if s[0] != 0 {
			f := 1
			if s[s[0]] != 0 {
				f = flip()
			}
			if f > maxflips {
				maxflips = f
			}
			if odd != 0 {
				checksum -= f
			} else {
				checksum += f
			}
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s number\n", os.Args[0])
		os.Exit(1)
	}

	max_n, _ = strconv.Atoi(os.Args[1])
	if max_n < 3 || max_n > 15 {
		fmt.Fprintf(os.Stderr, "range: must be 3 <= n <= 12\n")
		os.Exit(1)
	}

	for i := 0; i < max_n; i++ {
		s[i] = elem(i)
	}
	tk(max_n)

	fmt.Printf("%d\nPfannkuchen(%d) = %d\n", checksum, max_n, maxflips)
}

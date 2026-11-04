// binary-trees from the Computer Language Benchmarks Game, single threaded.
// https://benchmarksgame-team.pages.debian.net/benchmarksgame/description/binarytrees.html
// Every node is a separate allocation, so this measures the allocator and GC.
package main

import (
	"flag"
	"fmt"
	"strconv"
)

type node struct {
	left, right *node
}

func bottomUp(depth int) *node {
	if depth <= 0 {
		return &node{}
	}
	return &node{bottomUp(depth - 1), bottomUp(depth - 1)}
}

func (n *node) check() int {
	if n.left == nil {
		return 1
	}
	return 1 + n.left.check() + n.right.check()
}

func main() {
	flag.Parse()
	n, _ := strconv.Atoi(flag.Arg(0))
	minDepth := 4
	maxDepth := max(minDepth+2, n)

	stretch := maxDepth + 1
	fmt.Printf("stretch tree of depth %d\t check: %d\n", stretch, bottomUp(stretch).check())

	longLived := bottomUp(maxDepth)

	for depth := minDepth; depth <= maxDepth; depth += 2 {
		iterations := 1 << (maxDepth - depth + minDepth)
		check := 0
		for i := 0; i < iterations; i++ {
			check += bottomUp(depth).check()
		}
		fmt.Printf("%d\t trees of depth %d\t check: %d\n", iterations, depth, check)
	}
	fmt.Printf("long lived tree of depth %d\t check: %d\n", maxDepth, longLived.check())
}

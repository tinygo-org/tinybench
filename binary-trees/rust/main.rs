// binary-trees from the Computer Language Benchmarks Game, single threaded.
// https://benchmarksgame-team.pages.debian.net/benchmarksgame/description/binarytrees.html
// Every node is a separate Box allocation, to match the Go version.
use std::env;

struct Node {
    left: Option<Box<Node>>,
    right: Option<Box<Node>>,
}

fn bottom_up(depth: i32) -> Box<Node> {
    if depth <= 0 {
        return Box::new(Node { left: None, right: None });
    }
    Box::new(Node {
        left: Some(bottom_up(depth - 1)),
        right: Some(bottom_up(depth - 1)),
    })
}

fn check(n: &Node) -> i32 {
    match (&n.left, &n.right) {
        (Some(l), Some(r)) => 1 + check(l) + check(r),
        _ => 1,
    }
}

fn main() {
    let n: i32 = env::args().nth(1).and_then(|s| s.parse().ok()).unwrap_or(0);
    let min_depth = 4;
    let max_depth = std::cmp::max(min_depth + 2, n);

    let stretch = max_depth + 1;
    println!("stretch tree of depth {}\t check: {}", stretch, check(&bottom_up(stretch)));

    let long_lived = bottom_up(max_depth);

    let mut depth = min_depth;
    while depth <= max_depth {
        let iterations = 1 << (max_depth - depth + min_depth);
        let mut sum = 0;
        for _ in 0..iterations {
            sum += check(&bottom_up(depth));
        }
        println!("{}\t trees of depth {}\t check: {}", iterations, depth, sum);
        depth += 2;
    }
    println!("long lived tree of depth {}\t check: {}", max_depth, check(&long_lived));
}

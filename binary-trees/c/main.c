// binary-trees from the Computer Language Benchmarks Game, single threaded.
// https://benchmarksgame-team.pages.debian.net/benchmarksgame/description/binarytrees.html
// Every node is a separate malloc and free, to match the Go version.
#include <stdio.h>
#include <stdlib.h>

typedef struct node {
	struct node *left, *right;
} node;

static node *bottom_up(int depth) {
	node *n = malloc(sizeof(node));
	if (depth <= 0) {
		n->left = n->right = NULL;
	} else {
		n->left = bottom_up(depth - 1);
		n->right = bottom_up(depth - 1);
	}
	return n;
}

static int check(node *n) {
	if (n->left == NULL) {
		return 1;
	}
	return 1 + check(n->left) + check(n->right);
}

static void delete(node *n) {
	if (n->left != NULL) {
		delete(n->left);
		delete(n->right);
	}
	free(n);
}

int main(int argc, char **argv) {
	int n = argc > 1 ? atoi(argv[1]) : 0;
	int min_depth = 4;
	int max_depth = n > min_depth + 2 ? n : min_depth + 2;

	int stretch = max_depth + 1;
	node *t = bottom_up(stretch);
	printf("stretch tree of depth %d\t check: %d\n", stretch, check(t));
	delete(t);

	node *long_lived = bottom_up(max_depth);

	for (int depth = min_depth; depth <= max_depth; depth += 2) {
		int iterations = 1 << (max_depth - depth + min_depth);
		int sum = 0;
		for (int i = 0; i < iterations; i++) {
			t = bottom_up(depth);
			sum += check(t);
			delete(t);
		}
		printf("%d\t trees of depth %d\t check: %d\n", iterations, depth, sum);
	}
	printf("long lived tree of depth %d\t check: %d\n", max_depth, check(long_lived));
	delete(long_lived);
	return 0;
}

#include <stdio.h>
#include <stdlib.h>
#include <math.h>
#include <string.h>

static int n = 0;

static int evala(int i, int j) {
    return (i + j) * (i + j + 1) / 2 + i + 1;
}

static void times(double *v, const double *u, int n) {
    for (int i = 0; i < n; i++) {
        double a = 0.0;
        for (int j = 0; j < n; j++) {
            a += u[j] / (double)evala(i, j);
        }
        v[i] = a;
    }
}

static void times_trans(double *v, const double *u, int n) {
    for (int i = 0; i < n; i++) {
        double a = 0.0;
        for (int j = 0; j < n; j++) {
            a += u[j] / (double)evala(j, i);
        }
        v[i] = a;
    }
}

static void a_times_transp(double *v, const double *u, int n) {
    double *x = (double *)malloc(n * sizeof(double));
    times(x, u, n);
    times_trans(v, x, n);
    free(x);
}

int main(int argc, char *argv[]) {
    if (argc > 1) {
        n = atoi(argv[1]);
    }
    int verify = argc > 2 && strcmp(argv[2], "v") == 0;

    double *u = (double *)malloc(n * sizeof(double));
    double *v = (double *)malloc(n * sizeof(double));
    for (int i = 0; i < n; i++) {
        u[i] = 1.0;
        v[i] = 1.0;
    }

    for (int i = 0; i < 10; i++) {
        a_times_transp(v, u, n);
        a_times_transp(u, v, n);
    }

    double vBv = 0.0, vv = 0.0;
    for (int i = 0; i < n; i++) {
        vBv += u[i] * v[i];
        vv += v[i] * v[i];
    }

    double answer = sqrt(vBv / vv);
    if (verify) {
        printf("%0.9f\n", answer);
    }

    free(u);
    free(v);
    return 0;
}

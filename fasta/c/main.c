/* The Computer Language Benchmarks Game
 * https://salsa.debian.org/benchmarksgame-team/benchmarksgame/
 *
 * C version based on Go program by The Go Authors.
 * Based on C program by Joern Inge Vestgaarden
 * and Jorge Peixoto de Morais Neto.
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>

#define WIDTH 60

typedef struct {
    double p;
    char c;
} AminoAcid;

static int min(int a, int b) {
    return a < b ? a : b;
}

static void accumulate_probabilities(AminoAcid* genelist, int len) {
    for (int i = 1; i < len; i++) {
        genelist[i].p += genelist[i-1].p;
    }
}

static void repeat_fasta(const char* s, int slen, int count, FILE* out, int verify) {
    int pos = 0;
    char s2[WIDTH + slen];
    memcpy(s2, s, slen);
    memcpy(s2 + slen, s, WIDTH);
    while (count > 0) {
        int line = min(WIDTH, count);
        if (verify) {
            fwrite(s2 + pos, 1, line, out);
            fputc('\n', out);
        }
        pos += line;
        if (pos >= slen) pos -= slen;
        count -= line;
    }
}


static uint32_t lastrandom = 42;
#define IM 139968
#define IA 3877
#define IC 29573

static void random_fasta(AminoAcid* genelist, int len, int count, FILE* out, int verify) {
    char buf[WIDTH + 1];
    while (count > 0) {
        int line = min(WIDTH, count);
        for (int pos = 0; pos < line; pos++) {
            lastrandom = (lastrandom * IA + IC) % IM;
            double r = (double)((int)lastrandom) / IM;
            for (int i = 0; i < len; i++) {
                if (genelist[i].p >= r) {
                    buf[pos] = genelist[i].c;
                    break;
                }
            }
        }
        buf[line] = '\n';
        if (verify) {
            fwrite(buf, 1, line + 1, out);
        }
        count -= line;
    }
}

int main(int argc, char* argv[]) {
    int n = 0;
    if (argc > 1) {
        n = atoi(argv[1]);
    }
    int verify = argc > 2 && strcmp(argv[2], "v") == 0;

    AminoAcid iub[] = {
        {0.27, 'a'}, {0.12, 'c'}, {0.12, 'g'}, {0.27, 't'},
        {0.02, 'B'}, {0.02, 'D'}, {0.02, 'H'}, {0.02, 'K'},
        {0.02, 'M'}, {0.02, 'N'}, {0.02, 'R'}, {0.02, 'S'},
        {0.02, 'V'}, {0.02, 'W'}, {0.02, 'Y'}
    };
    int iub_len = sizeof(iub) / sizeof(iub[0]);

    AminoAcid homosapiens[] = {
        {0.3029549426680, 'a'},
        {0.1979883004921, 'c'},
        {0.1975473066391, 'g'},
        {0.3015094502008, 't'}
    };
    int homosapiens_len = sizeof(homosapiens) / sizeof(homosapiens[0]);

    accumulate_probabilities(iub, iub_len);
    accumulate_probabilities(homosapiens, homosapiens_len);

    const char* alu =
        "GGCCGGGCGCGGTGGCTCACGCCTGTAATCCCAGCACTTTGG"
        "GAGGCCGAGGCGGGCGGATCACCTGAGGTCAGGAGTTCGAGA"
        "CCAGCCTGGCCAACATGGTGAAACCCCGTCTCTACTAAAAAT"
        "ACAAAAATTAGCCGGGCGTGGTGGCGCGCGCCTGTAATCCCA"
        "GCTACTCGGGAGGCTGAGGCAGGAGAATCGCTTGAACCCGGG"
        "AGGCGGAGGTTGCAGTGAGCCGAGATCGCGCCACTGCACTCC"
        "AGCCTGGGCGACAGAGCGAGACTCCGTCTCAAAAA";

    if (verify) printf(">ONE Homo sapiens alu\n");
    repeat_fasta(alu, strlen(alu), 2 * n, stdout, verify);
    if (verify) printf(">TWO IUB ambiguity codes\n");
    random_fasta(iub, iub_len, 3 * n, stdout, verify);
    if (verify) printf(">THREE Homo sapiens frequency\n");
    random_fasta(homosapiens, homosapiens_len, 5 * n, stdout, verify);
    return 0;
}

/* The Computer Language Benchmarks Game
 * https://salsa.debian.org/benchmarksgame-team/benchmarksgame/
 *
 * Rust version based on Go program by The Go Authors.
 * Based on C program by Joern Inge Vestgaarden
 * and Jorge Peixoto de Morais Neto.
 */

use std::env;
use std::io::{self, Write};

const WIDTH: usize = 60;

#[derive(Clone, Copy)]
struct AminoAcid {
    p: f64,
    c: u8,
}

fn accumulate_probabilities(genelist: &mut [AminoAcid]) {
    for i in 1..genelist.len() {
        genelist[i].p += genelist[i - 1].p;
    }
}

fn repeat_fasta(s: &[u8], count: usize, out: &mut dyn Write) {
    let slen = s.len();
    let mut s2 = Vec::with_capacity(slen + WIDTH);
    s2.extend_from_slice(s);
    s2.extend_from_slice(&s[..WIDTH]);
    let mut pos = 0;
    let mut remaining = count;
    while remaining > 0 {
        let line = WIDTH.min(remaining);
        out.write_all(&s2[pos..pos + line]).unwrap();
        out.write_all(b"\n").unwrap();
        pos += line;
        if pos >= slen {
            pos -= slen;
        }
        remaining -= line;
    }
}

static mut LAST_RANDOM: u32 = 42;
const IM: u32 = 139968;
const IA: u32 = 3877;
const IC: u32 = 29573;

fn random_fasta(genelist: &[AminoAcid], count: usize, out: &mut dyn Write) {
    let mut buf = vec![0u8; WIDTH + 1];
    let mut remaining = count;
    while remaining > 0 {
        let line = WIDTH.min(remaining);
        for pos in 0..line {
            let lastrandom = unsafe {
                LAST_RANDOM = (LAST_RANDOM.wrapping_mul(IA).wrapping_add(IC)) % IM;
                LAST_RANDOM
            };
            let r = lastrandom as f64 / IM as f64;
            for aa in genelist {
                if aa.p >= r {
                    buf[pos] = aa.c;
                    break;
                }
            }
        }
        buf[line] = b'\n';
        out.write_all(&buf[..line + 1]).unwrap();
        remaining -= line;
    }
}

fn main() {
    let n = match env::args().nth(1) {
        Some(s) => s.parse::<usize>().unwrap_or(0),
        None => 0,
    };

    let mut iub = [
        AminoAcid { p: 0.27, c: b'a' },
        AminoAcid { p: 0.12, c: b'c' },
        AminoAcid { p: 0.12, c: b'g' },
        AminoAcid { p: 0.27, c: b't' },
        AminoAcid { p: 0.02, c: b'B' },
        AminoAcid { p: 0.02, c: b'D' },
        AminoAcid { p: 0.02, c: b'H' },
        AminoAcid { p: 0.02, c: b'K' },
        AminoAcid { p: 0.02, c: b'M' },
        AminoAcid { p: 0.02, c: b'N' },
        AminoAcid { p: 0.02, c: b'R' },
        AminoAcid { p: 0.02, c: b'S' },
        AminoAcid { p: 0.02, c: b'V' },
        AminoAcid { p: 0.02, c: b'W' },
        AminoAcid { p: 0.02, c: b'Y' },
    ];

    let mut homosapiens = [
        AminoAcid {
            p: 0.3029549426680,
            c: b'a',
        },
        AminoAcid {
            p: 0.1979883004921,
            c: b'c',
        },
        AminoAcid {
            p: 0.1975473066391,
            c: b'g',
        },
        AminoAcid {
            p: 0.3015094502008,
            c: b't',
        },
    ];

    accumulate_probabilities(&mut iub);
    accumulate_probabilities(&mut homosapiens);

    let alu = b"GGCCGGGCGCGGTGGCTCACGCCTGTAATCCCAGCACTTTGG\
GAGGCCGAGGCGGGCGGATCACCTGAGGTCAGGAGTTCGAGA\
CCAGCCTGGCCAACATGGTGAAACCCCGTCTCTACTAAAAAT\
ACAAAAATTAGCCGGGCGTGGTGGCGCGCGCCTGTAATCCCA\
GCTACTCGGGAGGCTGAGGCAGGAGAATCGCTTGAACCCGGG\
AGGCGGAGGTTGCAGTGAGCCGAGATCGCGCCACTGCACTCC\
AGCCTGGGCGACAGAGCGAGACTCCGTCTCAAAAA";

    let stdout = io::stdout();
    let mut out = stdout.lock();

    out.write_all(b">ONE Homo sapiens alu\n").unwrap();
    repeat_fasta(alu, 2 * n, &mut out);

    out.write_all(b">TWO IUB ambiguity codes\n").unwrap();
    random_fasta(&iub, 3 * n, &mut out);

    out.write_all(b">THREE Homo sapiens frequency\n").unwrap();
    random_fasta(&homosapiens, 5 * n, &mut out);
}

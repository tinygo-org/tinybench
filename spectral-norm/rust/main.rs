use std::env;

fn evala(i: usize, j: usize) -> usize {
    (i + j) * (i + j + 1) / 2 + i + 1
}

fn times(v: &mut [f64], u: &[f64]) {
    let n = v.len();
    for i in 0..n {
        let mut a = 0.0;
        for j in 0..n {
            a += u[j] / evala(i, j) as f64;
        }
        v[i] = a;
    }
}

fn times_trans(v: &mut [f64], u: &[f64]) {
    let n = v.len();
    for i in 0..n {
        let mut a = 0.0;
        for j in 0..n {
            a += u[j] / evala(j, i) as f64;
        }
        v[i] = a;
    }
}

fn a_times_transp(v: &mut [f64], u: &[f64]) {
    let n = u.len();
    let mut x = vec![0.0; n];
    times(&mut x, u);
    times_trans(v, &x);
}

fn main() {
    let mut args = env::args();
    let prog = args.next();
    let n: usize = match args.next() {
        Some(s) => s.parse().unwrap_or(0),
        None => 0,
    };

    let verify = args.next().as_deref() == Some("v");

    let mut u = vec![1.0; n];
    let mut v = vec![1.0; n];

    for _ in 0..10 {
        a_times_transp(&mut v, &u);
        a_times_transp(&mut u, &v);
    }

    let mut vBv = 0.0;
    let mut vv = 0.0;
    for i in 0..n {
        vBv += u[i] * v[i];
        vv += v[i] * v[i];
    }

    let answer = (vBv / vv).sqrt();
    if verify {
        println!("{:.9}", answer);
    }
}

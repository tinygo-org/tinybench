use std::env;
use std::process;

type Elem = u32;

#[derive(Debug, Default)]
struct Pfannkuch {
    s: [Elem; 16],
    t: [Elem; 16],
    maxflips: u32,
    max_n: u32,
    odd: u32,
    checksum: i32,
}

impl Pfannkuch {
    fn flip(&mut self, n_param: u32) -> u32 {
        for i in 0..(n_param as usize) {
            self.t[i] = self.s[i];
        }

        let mut flips_count = 1;

        loop {
            let mut x: usize = 0;
            let mut y: usize = self.t[0] as usize;

            while x < y {
                self.t.swap(x, y);
                x += 1;
                y -= 1;
            }

            flips_count += 1;

            if self.t[self.t[0] as usize] == 0 {
                break;
            }
        }

        flips_count
    }

    fn rotate(&mut self, n: u32) {
        let c = self.s[0];
        let n_usize = n as usize;
        for i in 0..n_usize {
            self.s[i] = self.s[i + 1];
        }
        self.s[n_usize] = c;
    }

    // n_param is self.max_n from main, which is the N for Fannkuch.
    fn tk(&mut self, n_param: u32) {
        let mut p_count = 0; // Permutation counter index, Go's 'i' in tk
        let mut c_perm_counts = [0 as Elem; 16]; // Permutation counts, Go's 'c' in tk

        while p_count < n_param {
            self.rotate(p_count);

            let p_count_usize = p_count as usize;

            if c_perm_counts[p_count_usize] >= p_count as Elem {
                c_perm_counts[p_count_usize] = 0;
                p_count += 1;
                continue;
            }

            c_perm_counts[p_count_usize] += 1;
            p_count = 1;
            self.odd = !self.odd;

            let idx = self.s[0] as usize;
            if idx != 0 {
                let mut f = 1;

                if self.s[idx] != 0 {
                    f = self.flip(n_param);
                }

                if f > self.maxflips {
                    self.maxflips = f;
                }

                if self.odd != 0 {
                    // not odd
                    self.checksum -= f as i32;
                } else {
                    // odd
                    self.checksum += f as i32;
                }
            }
        }
    }
}

fn main() {
    let mut args = env::args();

    let prog_name = args.next();

    let num_str = match args.next() {
        Some(number_str) => number_str,
        None => {
            // no number argument
            let name = prog_name.unwrap_or("fannkuch_redux_rust".to_string());
            eprintln!("usage: {name} number");
            process::exit(1);
        }
    };

    let mut pf = Pfannkuch::default(); // zero initialize by default

    match num_str.parse::<u32>() {
        Ok(n) => pf.max_n = n,
        Err(_) => {
            eprintln!("Error: '{num_str}' is not a valid number.");
            process::exit(1);
        }
    }

    if pf.max_n < 3 || pf.max_n > 15 {
        eprintln!("Error: N must be between 3 and 15, inclusive.");
        process::exit(1);
    }

    for i in 0..(pf.max_n as usize) {
        pf.s[i] = i as Elem;
    }

    pf.tk(pf.max_n);

    println!(
        "{}\nPfannkuchen({}) = {}",
        pf.checksum, pf.max_n, pf.maxflips
    );
}

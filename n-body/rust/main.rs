use std::env;
use std::f64::consts::PI;
use std::process;

const SOLAR_MASS: f64 = 4.0 * PI * PI;
const DAYS_PER_YEAR: f64 = 365.24;

#[derive(Clone, Copy, Debug)]
struct Planet {
    x: f64,
    y: f64,
    z: f64,
    vx: f64,
    vy: f64,
    vz: f64,
    mass: f64,
}

fn advance(bodies: &mut [Planet], dt: f64) {
    let n = bodies.len();
    for i in 0..n {
        for j in (i + 1)..n {
            let (body_i_slice, body_j_slice_plus) = bodies.split_at_mut(j);
            let b1 = &mut body_i_slice[i];
            let b2 = &mut body_j_slice_plus[0]; // This is bodies[j]

            let dx = b1.x - b2.x;
            let dy = b1.y - b2.y;
            let dz = b1.z - b2.z;

            let dist_sq = dx * dx + dy * dy + dz * dz;
            let dist = dist_sq.sqrt();
            let mag = dt / (dist_sq * dist);

            let mass_b2_mag = b2.mass * mag;
            b1.vx -= dx * mass_b2_mag;
            b1.vy -= dy * mass_b2_mag;
            b1.vz -= dz * mass_b2_mag;

            let mass_b1_mag = b1.mass * mag;
            b2.vx += dx * mass_b1_mag;
            b2.vy += dy * mass_b1_mag;
            b2.vz += dz * mass_b1_mag;
        }
    }

    for i in 0..n {
        bodies[i].x += dt * bodies[i].vx;
        bodies[i].y += dt * bodies[i].vy;
        bodies[i].z += dt * bodies[i].vz;
    }
}

fn energy(bodies: &[Planet]) -> f64 {
    let mut e = 0.0;
    let n = bodies.len();
    for i in 0..n {
        let b1 = &bodies[i];
        e += 0.5 * b1.mass * (b1.vx * b1.vx + b1.vy * b1.vy + b1.vz * b1.vz);
        for j in (i + 1)..n {
            let b2 = &bodies[j];
            let dx = b1.x - b2.x;
            let dy = b1.y - b2.y;
            let dz = b1.z - b2.z;
            let distance = (dx * dx + dy * dy + dz * dz).sqrt();
            e -= (b1.mass * b2.mass) / distance;
        }
    }
    e
}

fn offset_momentum(bodies: &mut [Planet]) {
    let mut px = 0.0;
    let mut py = 0.0;
    let mut pz = 0.0;
    for body in bodies.iter() {
        px += body.vx * body.mass;
        py += body.vy * body.mass;
        pz += body.vz * body.mass;
    }

    bodies[0].vx = -px / SOLAR_MASS;
    bodies[0].vy = -py / SOLAR_MASS;
    bodies[0].vz = -pz / SOLAR_MASS;
}

const NBODIES: usize = 5;

fn initial_bodies() -> [Planet; NBODIES] {
    [
        Planet { // sun
            x: 0.0, y: 0.0, z: 0.0, vx: 0.0, vy: 0.0, vz: 0.0, mass: SOLAR_MASS,
        },
        Planet { // jupiter
            x: 4.84143144246472090e+00,
            y: -1.16032004402742839e+00,
            z: -1.03622044471123109e-01,
            vx: 1.66007664274403694e-03 * DAYS_PER_YEAR,
            vy: 7.69901118419740425e-03 * DAYS_PER_YEAR,
            vz: -6.90460016972063023e-05 * DAYS_PER_YEAR,
            mass: 9.54791938424326609e-04 * SOLAR_MASS,
        },
        Planet { // saturn
            x: 8.34336671824457987e+00,
            y: 4.12479856412430479e+00,
            z: -4.03523417114321381e-01,
            vx: -2.76742510726862411e-03 * DAYS_PER_YEAR,
            vy: 4.99852801234917238e-03 * DAYS_PER_YEAR,
            vz: 2.30417297573763929e-05 * DAYS_PER_YEAR,
            mass: 2.85885980666130812e-04 * SOLAR_MASS,
        },
        Planet { // uranus
            x: 1.28943695621391310e+01,
            y: -1.51111514016986312e+01,
            z: -2.23307578892655734e-01,
            vx: 2.96460137564761618e-03 * DAYS_PER_YEAR,
            vy: 2.37847173959480950e-03 * DAYS_PER_YEAR,
            vz: -2.96589568540237556e-05 * DAYS_PER_YEAR,
            mass: 4.36624404335156298e-05 * SOLAR_MASS,
        },
        Planet { // neptune
            x: 1.53796971148509165e+01,
            y: -2.59193146099879641e+01,
            z: 1.79258772950371181e-01,
            vx: 2.68067772490389322e-03 * DAYS_PER_YEAR,
            vy: 1.62824170038242295e-03 * DAYS_PER_YEAR,
            vz: -9.51592254519715870e-05 * DAYS_PER_YEAR,
            mass: 5.15138902046611451e-05 * SOLAR_MASS,
        },
    ]
}

fn main() {
    let args: Vec<String> = env::args().collect();
    if args.len() < 2 {
        eprintln!("Usage: {} <number_of_steps>", args.get(0).map_or("nbody_rust", |s| s.as_str()));
        process::exit(1);
    }

    let n_steps: usize = match args[1].parse() {
        Ok(n) => n,
        Err(_) => {
            eprintln!("Error: Could not parse number of steps '{}'", args[1]);
            process::exit(1);
        }
    };

    let mut bodies_arr = initial_bodies();
    offset_momentum(&mut bodies_arr);

    println!("{:.9}", energy(&bodies_arr));

    for _ in 0..n_steps {
        advance(&mut bodies_arr, 0.01);
    }

    println!("{:.9}", energy(&bodies_arr));
}

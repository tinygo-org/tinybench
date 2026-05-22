// Original Author: Juan Pablo Sellanes Goncalves
// https://eventos.iua.edu.ar/event/1/contributions/51/

use std::env;
use std::f64::consts::PI;
use std::process;

// --- Physical constants ---
const G0: f64 = 9.807e-3;
const DAYS: f64 = 86400.0;
const BIG_G: f64 = 6.6742e-20;
const R12: f64 = 384400.0;
const M1: f64 = 5.974e24;
const M2: f64 = 7.348e22;
const MU1: f64 = 398600.0;
const MU2: f64 = 4903.02;
const MS: f64 = 1.989e30;
const RB2S: f64 = 149597870.7;
const R_EARTH: f64 = 6378.0;
const L1DIST: f64 = 321710.0;
const C1: f64 = -1.676;

// --- Vec3 (equivalent to Go's Vec struct) ---
#[derive(Clone, Copy, Default)]
struct Vec3 {
    x: f64,
    y: f64,
    z: f64,
}

impl Vec3 {
    fn norm2(self) -> f64 { self.x * self.x + self.y * self.y + self.z * self.z }
    fn norm(self) -> f64 { self.x.hypot(self.y.hypot(self.z)) }
    fn add(self, b: Vec3) -> Vec3 { Vec3 { x: self.x + b.x, y: self.y + b.y, z: self.z + b.z } }
    fn sub(self, b: Vec3) -> Vec3 { Vec3 { x: self.x - b.x, y: self.y - b.y, z: self.z - b.z } }
    fn scale(self, f: f64) -> Vec3 { Vec3 { x: f * self.x, y: f * self.y, z: f * self.z } }
}

#[allow(dead_code)]
fn v3_fma(a: f64, b: Vec3, c: Vec3) -> Vec3 {
    Vec3 { x: a.mul_add(b.x, c.x), y: a.mul_add(b.y, c.y), z: a.mul_add(b.z, c.z) }
}

// --- State ---
#[derive(Clone, Copy)]
struct State {
    pos: Vec3,
    vel: Vec3,
    mass: f64,
}

// --- Thruster ---
#[derive(Clone, Copy)]
struct Thruster {
    thrust: f64,
    isp: f64,
}

impl Thruster {
    fn mass_rate(self) -> f64 { -self.thrust / (G0 * self.isp) }
}

// --- GtoConfig ---
struct GtoConfig {
    h_apogee: f64,
    h_perigee: f64,
    phi: f64,
    gamma: f64,
    m0: f64,
    thruster: Thruster,
    jacobi_thr: f64,
    tol: f64,
}

impl GtoConfig {
    fn gto_initial_state(&self, pc: PhysConsts) -> State {
        let r_apogee = R_EARTH + self.h_apogee;
        let r_perigee = R_EARTH + self.h_perigee;
        let e = (r_apogee - r_perigee) / (r_apogee + r_perigee);
        let v0 = (MU1 * (1.0 - e) / r_apogee).sqrt() - pc.w * r_apogee;

        let sin_phi = self.phi.sin();
        let cos_phi = self.phi.cos();
        let sin_gam = self.gamma.sin();
        let cos_gam = self.gamma.cos();

        State {
            pos: Vec3 { x: r_apogee * cos_phi + pc.x1, y: r_apogee * sin_phi, z: 0.0 },
            vel: Vec3 {
                x: v0 * (sin_gam * cos_phi - cos_gam * sin_phi),
                y: v0 * (sin_gam * sin_phi + cos_gam * cos_phi),
                z: 0.0,
            },
            mass: self.m0,
        }
    }
}

// --- PhysConsts: runtime-derived globals (Go uses package-level var for these). ---
#[derive(Clone, Copy)]
struct PhysConsts {
    w: f64,  // [rad/s] angular velocity of rotating frame
    x1: f64, // [km] Earth x-position in rotating frame
    x2: f64, // [km] Moon x-position in rotating frame
}

impl PhysConsts {
    fn new() -> Self {
        PhysConsts {
            w: ((MU1 + MU2) / (R12 * R12 * R12)).sqrt(),
            x1: -(M2 / (M1 + M2)) * R12,
            x2: (M1 / (M1 + M2)) * R12,
        }
    }
}

// --- gravity ---
fn gravity(x: f64, y: f64, pc: PhysConsts) -> (f64, f64) {
    let dx1 = x - pc.x1;
    let dx2 = x - pc.x2;
    let r1 = (dx1 * dx1 + y * y).sqrt();
    let r2 = (dx2 * dx2 + y * y).sqrt();
    (r1 * r1 * r1, r2 * r2 * r2)
}

// --- Rates functions ---

fn rates_thrust_em(_t: f64, s: State, _phi_s0: f64, th: Thruster, pc: PhysConsts) -> (Vec3, Vec3, f64) {
    let (x, y) = (s.pos.x, s.pos.y);
    let (vx, vy, m) = (s.vel.x, s.vel.y, s.mass);
    let (r1_3, r2_3) = gravity(x, y, pc);
    let v = (vx * vx + vy * vy).sqrt();
    let tmv = th.thrust / (m * v);
    let ax = ((2.0 * pc.w * vy + pc.w * pc.w * x) - MU1 * (x - pc.x1) / r1_3)
        - MU2 * (x - pc.x2) / r2_3 + tmv * vx;
    let ay = (-2.0 * pc.w * vx + pc.w * pc.w * y) - (MU1 / r1_3 + MU2 / r2_3) * y + tmv * vy;
    (s.vel, Vec3 { x: ax, y: ay, z: 0.0 }, th.mass_rate())
}

fn rates_coast_em(_t: f64, s: State, _phi_s0: f64, pc: PhysConsts) -> (Vec3, Vec3, f64) {
    let (x, y) = (s.pos.x, s.pos.y);
    let (vx, vy) = (s.vel.x, s.vel.y);
    let (r1_3, r2_3) = gravity(x, y, pc);
    let ax = ((2.0 * pc.w * vy + pc.w * pc.w * x) - MU1 * (x - pc.x1) / r1_3)
        - MU2 * (x - pc.x2) / r2_3;
    let ay = (-2.0 * pc.w * vx + pc.w * pc.w * y) - (MU1 / r1_3 + MU2 / r2_3) * y;
    (s.vel, Vec3 { x: ax, y: ay, z: 0.0 }, 0.0)
}

fn rates_brake_em(_t: f64, s: State, _phi_s0: f64, th: Thruster, pc: PhysConsts) -> (Vec3, Vec3, f64) {
    let (x, y) = (s.pos.x, s.pos.y);
    let (vx, vy, m) = (s.vel.x, s.vel.y, s.mass);
    let (r1_3, r2_3) = gravity(x, y, pc);
    let v = (vx * vx + vy * vy).sqrt();
    let tmv = -th.thrust / (m * v); // negative: retrograde
    let ax = ((2.0 * pc.w * vy + pc.w * pc.w * x) - MU1 * (x - pc.x1) / r1_3)
        - MU2 * (x - pc.x2) / r2_3 + tmv * vx;
    let ay = (-2.0 * pc.w * vx + pc.w * pc.w * y) - (MU1 / r1_3 + MU2 / r2_3) * y + tmv * vy;
    (s.vel, Vec3 { x: ax, y: ay, z: 0.0 }, th.mass_rate())
}

// RatesMode: Rust enum equivalent of Go's RatesFunc closures.
// Each variant captures the data its Go closure would have captured.
#[derive(Clone, Copy)]
enum RatesMode {
    Thrust(Thruster),
    Coast,
    Brake(Thruster),
}

impl RatesMode {
    fn compute(&self, t: f64, s: State, phi_s0: f64, pc: PhysConsts) -> (Vec3, Vec3, f64) {
        match self {
            RatesMode::Thrust(th) => rates_thrust_em(t, s, phi_s0, *th, pc),
            RatesMode::Coast => rates_coast_em(t, s, phi_s0, pc),
            RatesMode::Brake(th) => rates_brake_em(t, s, phi_s0, *th, pc),
        }
    }
}

// --- Jacobi constant and events ---

fn jacobi_constant_em(s: State, pc: PhysConsts) -> f64 {
    let v2 = s.vel.norm2();
    let d1 = (s.pos.x - pc.x1).hypot(s.pos.y);
    let d2 = (s.pos.x - pc.x2).hypot(s.pos.y);
    0.5 * v2 - 0.5 * pc.w * pc.w * (s.pos.x * s.pos.x + s.pos.y * s.pos.y) - MU1 / d1 - MU2 / d2
}

// Event: Rust enum equivalent of Go's EventFunc closures.
// Each variant captures the data its Go closure would have captured.
#[derive(Clone, Copy)]
enum Event {
    JacobiEM(f64), // captures threshold
    L1Em,
}

impl Event {
    fn eval(&self, _t: f64, s: State, _phi_s0: f64, pc: PhysConsts) -> f64 {
        match self {
            Event::JacobiEM(thr) => jacobi_constant_em(s, pc) - thr,
            Event::L1Em => (s.pos.x - pc.x1).hypot(s.pos.y) - L1DIST,
        }
    }
}

// --- Integrator ---
struct Integrator {
    t: f64,
    state: State,
    phi_s0: f64,
    rates: RatesMode,
    min_step: f64,
    max_step: f64,
    atol: f64,
    rtol: f64,
    last_err_norm: f64,
    step_count: i64,
    pc: PhysConsts,
}

// --- Dormand-Prince RK45 coefficients ---
const DP_C: [f64; 7] = [0.0, 1.0 / 5.0, 3.0 / 10.0, 4.0 / 5.0, 8.0 / 9.0, 1.0, 1.0];

const DP_A: [[f64; 6]; 7] = [
    [0.0, 0.0, 0.0, 0.0, 0.0, 0.0],
    [1.0 / 5.0, 0.0, 0.0, 0.0, 0.0, 0.0],
    [3.0 / 40.0, 9.0 / 40.0, 0.0, 0.0, 0.0, 0.0],
    [44.0 / 45.0, -56.0 / 15.0, 32.0 / 9.0, 0.0, 0.0, 0.0],
    [19372.0 / 6561.0, -25360.0 / 2187.0, 64448.0 / 6561.0, -212.0 / 729.0, 0.0, 0.0],
    [9017.0 / 3168.0, -355.0 / 33.0, 46732.0 / 5247.0, 49.0 / 176.0, -5103.0 / 18656.0, 0.0],
    [35.0 / 384.0, 0.0, 500.0 / 1113.0, 125.0 / 192.0, -2187.0 / 6784.0, 11.0 / 84.0],
];

const DP_B: [f64; 7] = [
    35.0 / 384.0, 0.0, 500.0 / 1113.0, 125.0 / 192.0, -2187.0 / 6784.0, 11.0 / 84.0, 0.0,
];

const DP_E: [f64; 7] = [
    71.0 / 57600.0,
    0.0,
    -71.0 / 16695.0,
    71.0 / 1920.0,
    -17253.0 / 339200.0,
    22.0 / 525.0,
    -1.0 / 40.0,
];

impl Integrator {
    #[allow(dead_code)]
    fn select_initial_step(&self) -> f64 {
        const ERR_ORDER: f64 = 4.0;
        const N_COMP: f64 = 5.0;

        let s = self.state;
        let pc = self.pc;
        let (d_pos0, d_vel0, dm0) = self.rates.compute(self.t, s, self.phi_s0, pc);

        let atol = self.atol;
        let rtol = self.rtol;
        let sc = [
            atol + s.pos.x.abs() * rtol,
            atol + s.pos.y.abs() * rtol,
            atol + s.pos.z.abs() * rtol,
            atol + s.vel.x.abs() * rtol,
            atol + s.vel.y.abs() * rtol,
            atol + s.vel.z.abs() * rtol,
            atol + s.mass.abs() * rtol,
        ];
        let rms_norm = |px: f64, py: f64, pz: f64, vx: f64, vy: f64, vz: f64, m: f64| {
            ((px * px + py * py + pz * pz + vx * vx + vy * vy + vz * vz + m * m) / N_COMP).sqrt()
        };

        let d0 = rms_norm(
            s.pos.x / sc[0], s.pos.y / sc[1], s.pos.z / sc[2],
            s.vel.x / sc[3], s.vel.y / sc[4], s.vel.z / sc[5], s.mass / sc[6],
        );
        let d1 = rms_norm(
            d_pos0.x / sc[0], d_pos0.y / sc[1], d_pos0.z / sc[2],
            d_vel0.x / sc[3], d_vel0.y / sc[4], d_vel0.z / sc[5], dm0 / sc[6],
        );

        let h0 = if d0 < 1e-5 || d1 < 1e-5 { 1e-6 } else { 0.01 * d0 / d1 };

        let s1 = State {
            pos: s.pos.add(d_pos0.scale(h0)),
            vel: s.vel.add(d_vel0.scale(h0)),
            mass: s.mass + h0 * dm0,
        };
        let (d_pos1, d_vel1, dm1) = self.rates.compute(self.t + h0, s1, self.phi_s0, pc);
        let dd_pos = d_pos1.sub(d_pos0);
        let dd_vel = d_vel1.sub(d_vel0);
        let ddm = dm1 - dm0;
        let d2 = rms_norm(
            dd_pos.x / sc[0], dd_pos.y / sc[1], dd_pos.z / sc[2],
            dd_vel.x / sc[3], dd_vel.y / sc[4], dd_vel.z / sc[5], ddm / sc[6],
        ) / h0;

        let max_d = d1.max(d2);
        let h1 = if max_d <= 1e-5 {
            (h0 * 1e-3).max(1e-6)
        } else {
            (0.01 / max_d).powf(1.0 / (ERR_ORDER + 1.0))
        };

        (100.0 * h0).min(h1)
    }

    fn step(&mut self, h: f64) -> f64 {
        const SAFETY: f64 = 0.9;
        const MIN_FAC: f64 = 0.2;
        const MAX_FAC: f64 = 10.0;
        const ORDER: f64 = 5.0;

        // Copy read-only fields before the loop to avoid borrow conflicts.
        let t = self.t;
        let s = self.state;
        let phi_s0 = self.phi_s0;
        let rates = self.rates;
        let atol = self.atol;
        let rtol = self.rtol;
        let min_step = self.min_step;
        let max_step = self.max_step;
        let pc = self.pc;

        let mut h = h;
        let mut step_rejected = false;
        loop {
            let t_next = t + h;
            let h_eff = t_next - t;

            let mut k_pos = [Vec3::default(); 7];
            let mut k_vel = [Vec3::default(); 7];
            let mut k_m = [0.0f64; 7];

            let (dp, dv, dm) = rates.compute(t, s, phi_s0, pc);
            k_pos[0] = dp; k_vel[0] = dv; k_m[0] = dm;

            for i in 1..7 {
                let mut sum_pos = Vec3::default();
                let mut sum_vel = Vec3::default();
                let mut sum_mass = 0.0f64;
                for j in 0..i {
                    sum_pos.x += DP_A[i][j] * k_pos[j].x;
                    sum_pos.y += DP_A[i][j] * k_pos[j].y;
                    sum_vel.x += DP_A[i][j] * k_vel[j].x;
                    sum_vel.y += DP_A[i][j] * k_vel[j].y;
                    sum_mass  += DP_A[i][j] * k_m[j];
                }
                let ti = t + DP_C[i] * h_eff;
                let si = State {
                    pos:  s.pos.add(sum_pos.scale(h_eff)),
                    vel:  s.vel.add(sum_vel.scale(h_eff)),
                    mass: s.mass + h_eff * sum_mass,
                };
                let (dp, dv, dm) = rates.compute(ti, si, phi_s0, pc);
                k_pos[i] = dp; k_vel[i] = dv; k_m[i] = dm;
            }

            let mut sum_b_pos = Vec3::default();
            let mut sum_b_vel = Vec3::default();
            let mut sum_e_pos = Vec3::default();
            let mut sum_e_vel = Vec3::default();
            let mut sum_bm = 0.0f64;
            let mut sum_em = 0.0f64;
            for i in 0..7 {
                sum_b_pos.x += DP_B[i] * k_pos[i].x;
                sum_b_pos.y += DP_B[i] * k_pos[i].y;
                sum_b_vel.x += DP_B[i] * k_vel[i].x;
                sum_b_vel.y += DP_B[i] * k_vel[i].y;
                sum_bm      += DP_B[i] * k_m[i];
                sum_e_pos.x += DP_E[i] * k_pos[i].x;
                sum_e_pos.y += DP_E[i] * k_pos[i].y;
                sum_e_vel.x += DP_E[i] * k_vel[i].x;
                sum_e_vel.y += DP_E[i] * k_vel[i].y;
                sum_em      += DP_E[i] * k_m[i];
            }

            let new_pos = s.pos.add(sum_b_pos.scale(h_eff));
            let new_vel = s.vel.add(sum_b_vel.scale(h_eff));
            let new_m   = s.mass + h_eff * sum_bm;
            let err_pos = sum_e_pos.scale(h_eff);
            let err_vel = sum_e_vel.scale(h_eff);
            let err_m   = h_eff * sum_em;

            // Per-component error scaling, n=5 state components (x,y,vx,vy,m).
            let sc_px = atol + rtol * s.pos.x.abs().max(new_pos.x.abs());
            let sc_py = atol + rtol * s.pos.y.abs().max(new_pos.y.abs());
            let sc_vx = atol + rtol * s.vel.x.abs().max(new_vel.x.abs());
            let sc_vy = atol + rtol * s.vel.y.abs().max(new_vel.y.abs());
            let sc_m  = atol + rtol * s.mass.abs().max(new_m.abs());

            let sum_sq = (err_pos.x / sc_px).powi(2)
                + (err_pos.y / sc_py).powi(2)
                + (err_vel.x / sc_vx).powi(2)
                + (err_vel.y / sc_vy).powi(2)
                + (err_m / sc_m).powi(2);
            let err_norm = sum_sq.sqrt() / 5.0_f64.sqrt();

            if err_norm <= 1.0 {
                self.t = t_next;
                self.state = State { pos: new_pos, vel: new_vel, mass: new_m };
                self.last_err_norm = err_norm;
                self.step_count += 1;

                if err_norm == 0.0 {
                    return (h_eff * MAX_FAC).min(max_step);
                }
                let factor_raw = SAFETY * err_norm.powf(-1.0 / ORDER);
                let mut factor = factor_raw.clamp(MIN_FAC, MAX_FAC);
                if step_rejected { factor = factor.min(1.0); }
                return (h_eff * factor).clamp(min_step, max_step);
            }

            // Step rejected: reduce h.
            let factor = (SAFETY * err_norm.powf(-1.0 / ORDER)).max(MIN_FAC);
            h = (h_eff * factor).max(min_step);
            step_rejected = true;

            if h <= min_step {
                self.t = t_next;
                self.state = State { pos: new_pos, vel: new_vel, mass: new_m };
                return min_step;
            }
        }
    }
}

// --- bisect_event ---
fn bisect_event(
    mut t0: f64, mut t1: f64,
    mut s0: State, mut s1: State,
    phi_s0: f64, ev: &Event, pc: PhysConsts,
) -> f64 {
    const MAX_ITER: i32 = 50;
    const TOL: f64 = 1e-10;

    for _ in 0..MAX_ITER {
        let tm = 0.5 * (t0 + t1);
        if t1 - t0 < TOL {
            return tm;
        }
        let alpha = (tm - t0) / (t1 - t0);
        let sm = State {
            pos:  s0.pos.scale(1.0 - alpha).add(s1.pos.scale(alpha)),
            vel:  s0.vel.scale(1.0 - alpha).add(s1.vel.scale(alpha)),
            mass: (1.0 - alpha) * s0.mass + alpha * s1.mass,
        };
        let vm = ev.eval(tm, sm, phi_s0, pc);
        let v0 = ev.eval(t0, s0, phi_s0, pc);

        if v0 * vm < 0.0 {
            t1 = tm; s1 = sm;
        } else {
            t0 = tm; s0 = sm;
        }
    }
    0.5 * (t0 + t1)
}

// --- integrate_until ---
// Equivalent to Go's IntegrateUntil with variadic EventFunc; events is a slice here.
fn integrate_until(ig: &mut Integrator, tf: f64, events: &[Event]) -> (i32, f64) {
    let mut h = ig.max_step;
    let pc = ig.pc;

    let mut prev_ev = [0.0f64; 4]; // max 4 events; benchmark uses at most 1
    for (i, ev) in events.iter().enumerate() {
        prev_ev[i] = ev.eval(ig.t, ig.state, ig.phi_s0, pc);
    }

    while ig.t < tf {
        if ig.t + h > tf {
            h = tf - ig.t;
        }
        let t_prev = ig.t;
        let s_prev = ig.state;
        h = ig.step(h);

        for (i, ev) in events.iter().enumerate() {
            let curr = ev.eval(ig.t, ig.state, ig.phi_s0, pc);
            if prev_ev[i] * curr < 0.0 {
                let event_time =
                    bisect_event(t_prev, ig.t, s_prev, ig.state, ig.phi_s0, ev, pc);
                let alpha = (event_time - t_prev) / (ig.t - t_prev);
                let cur_state = ig.state;
                ig.state.pos  = s_prev.pos.scale(1.0 - alpha).add(cur_state.pos.scale(alpha));
                ig.state.vel  = s_prev.vel.scale(1.0 - alpha).add(cur_state.vel.scale(alpha));
                ig.state.mass = (1.0 - alpha) * s_prev.mass + alpha * cur_state.mass;
                ig.t = event_time;
                return (i as i32, event_time);
            }
            prev_ev[i] = curr;
        }
    }
    (-1, ig.t)
}

// --- print_state ---
fn print_state(phase: i32, t: f64, s: State) {
    println!(
        "phase={} t={:6.2}d earthdist={:.1}km pos=({:.3},{:.3})",
        phase, t / DAYS, s.pos.norm(), s.pos.x, s.pos.y
    );
}

// --- run ---
fn run(lower_bound_step: f64, verify: bool, pc: PhysConsts) -> Result<(), &'static str> {
    let cfg = GtoConfig {
        h_apogee:   37000.0,
        h_perigee:  1200.0,
        phi:        0.7505211952744961 * PI / 180.0,
        gamma:      0.0,
        m0:         12.0,
        thruster:   Thruster { thrust: 4.0 * 0.000000450, isp: 1650.0 },
        jacobi_thr: -1.63907788,
        tol:        1e-12,
    };

    let mut ig = Integrator {
        t:             0.0,
        state:         cfg.gto_initial_state(pc),
        phi_s0:        0.0,
        rates:         RatesMode::Thrust(cfg.thruster),
        min_step:      0.0,
        max_step:      450.0_f64.min(lower_bound_step),
        atol:          cfg.tol,
        rtol:          1e-9,
        last_err_norm: 0.0,
        step_count:    0,
        pc,
    };

    if verify { print_state(0, 0.0, ig.state); }

    // Phase 1: prograde thrust until Jacobi threshold.
    let ev1 = [Event::JacobiEM(cfg.jacobi_thr)];
    let (ev_idx, t1) = integrate_until(&mut ig, DAYS * 360.0, &ev1);
    if ev_idx < 0 {
        return Err("phase 1: jacobi threshold not crossed");
    }
    if verify { print_state(1, t1, ig.state); }

    // Phase 2: coast until distance from earth = L1.
    ig.rates    = RatesMode::Coast;
    ig.max_step = 200.0_f64.min(lower_bound_step);
    let ev2 = [Event::L1Em];
    let (ev_idx, t2) = integrate_until(&mut ig, t1 + 260.0 * DAYS, &ev2);
    if ev_idx < 0 {
        return Err("during phase 2: L1 distance not reached");
    }
    if verify { print_state(2, t2, ig.state); }

    // Phase 3: retrograde braking until Jacobi constant reaches C1.
    ig.rates    = RatesMode::Brake(cfg.thruster);
    ig.max_step = (1.0 * DAYS).min(lower_bound_step);
    let ev3 = [Event::JacobiEM(C1)];
    let (ev_idx, t3) = integrate_until(&mut ig, t2 + 25.0 * DAYS, &ev3);
    if ev_idx < 0 {
        return Err("during phase 3: C1 distance not reached while braking");
    }
    if verify { print_state(3, t3, ig.state); }

    // Phase 4: 20-day coast, no events.
    ig.rates    = RatesMode::Coast;
    ig.max_step = 100.0_f64.min(lower_bound_step);
    let (_, t4) = integrate_until(&mut ig, t3 + 20.0 * DAYS, &[]);
    if verify {
        print_state(4, t4, ig.state);
        println!("num steps={}", ig.step_count);
    }

    Ok(())
}

fn main() {
    let mut args = env::args();
    let prog_name = args.next();
    let step_arg  = args.next();
    let verify    = args.next().as_deref() == Some("v");

    let lower_bound_step: f64 = match step_arg {
        None => {
            let name = prog_name.as_deref().unwrap_or("gto-lunar");
            eprintln!("Usage: {name} <lower_bound_step>");
            process::exit(1);
        }
        Some(s) => match s.parse() {
            Ok(v) => v,
            Err(_) => {
                eprintln!("bad arg: {s}");
                process::exit(1);
            }
        },
    };

    let pc = PhysConsts::new();
    let _ns = ((BIG_G * MS) / (RB2S * RB2S * RB2S)).sqrt(); // mirrors Go's var nS

    if let Err(e) = run(lower_bound_step, verify, pc) {
        eprintln!("{e}");
        process::exit(1);
    }
}

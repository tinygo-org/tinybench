// Original Author: Juan Pablo Sellanes Goncalves
// https://eventos.iua.edu.ar/event/1/contributions/51/

const std = @import("std");
const math = std.math;
const stdout = std.debug;

// Physical constants, mostly relating to Earth-moon-sun system.
const g0: f64 = 9.807e-3; // [km/s²]
const days: f64 = 24 * 3600; // [s/day]
const big_g: f64 = 6.6742e-20; // [km³/kg/s²]
const r12: f64 = 384400.0; // [km] Earth-Moon distance
const m1: f64 = 5.974e24; // [kg] Earth mass
const m2: f64 = 7.348e22; // [kg] Moon mass
const mu1: f64 = 398600.0; // [km³/s²] Earth gravitational parameter
const mu2: f64 = 4903.02; // [km³/s²] Moon gravitational parameter
const ms: f64 = 1.989e30; // [kg] Sun mass
const rb2s: f64 = 149597870.7; // [km] Barycenter to Sun distance
const r_moon: f64 = 1737.0; // [km] Moon radius
const r_earth: f64 = 6378.0; // [km] Earth radius
const l1dist: f64 = 321710.0; // [km] L1 distance from Earth center
const c1: f64 = -1.676; // Jacobi constant for braking phase

const mu_em: f64 = mu1 + mu2; // [km³/s²]
const pi1: f64 = m1 / (m1 + m2); // Earth mass fraction
const pi2: f64 = m2 / (m1 + m2); // Moon mass fraction
const x1: f64 = -pi2 * r12; // [km] Earth x-position in rotating frame
const x2: f64 = pi1 * r12; // [km] Moon x-position in rotating frame
const mu_sun: f64 = big_g * ms; // [km³/s²] Sun gravitational parameter

// Runtime-derived constants (Go uses package-level var).
const n_s: f64 = math.sqrt(mu_sun / (rb2s * rb2s * rb2s)); // [rad/s] Sun mean motion
const W: f64 = math.sqrt(mu_em / (r12 * r12 * r12)); // [rad/s] Angular velocity of rotating frame

const Vec = struct {
    x: f64 = 0,
    y: f64 = 0,
    z: f64 = 0,
};

fn v3Norm2(v: Vec) f64 {
    return v.x * v.x + v.y * v.y + v.z * v.z;
}

fn v3Norm(v: Vec) f64 {
    return math.hypot(v.x, math.hypot(v.y, v.z));
}

fn v3Sub(a: Vec, b: Vec) Vec {
    return Vec{ .x = a.x - b.x, .y = a.y - b.y, .z = a.z - b.z };
}

fn v3Add(a: Vec, b: Vec) Vec {
    return Vec{ .x = a.x + b.x, .y = a.y + b.y, .z = a.z + b.z };
}

fn v3Scale(f: f64, v: Vec) Vec {
    return Vec{ .x = f * v.x, .y = f * v.y, .z = f * v.z };
}

fn v3FMA(a: f64, b: Vec, c: Vec) Vec {
    return Vec{
        .x = math.fma(f64, a, b.x, c.x),
        .y = math.fma(f64, a, b.y, c.y),
        .z = math.fma(f64, a, b.z, c.z),
    };
}

const State = struct {
    pos: Vec,
    vel: Vec,
    mass: f64,
};

const Thruster = struct {
    thrust: f64,
    isp: f64,

    fn massRate(th: Thruster) f64 {
        return -th.thrust / (g0 * th.isp);
    }
};

// RatesResult holds the three return values of a rates function.
const RatesResult = struct {
    d_pos: Vec,
    d_vel: Vec,
    dm: f64,
};

// RatesMode is the Zig equivalent of Go's RatesFunc closures.
const RatesMode = union(enum) {
    thrust: Thruster,
    coast: void,
    brake: Thruster,

    fn compute(mode: RatesMode, t: f64, s: State, phi_s0: f64) RatesResult {
        _ = t;
        _ = phi_s0;
        return switch (mode) {
            .thrust => |th| ratesThrustEM(s, th),
            .coast => ratesCoastEM(s),
            .brake => |th| ratesBrakeEM(s, th),
        };
    }
};

fn ratesThrustEM(s: State, th: Thruster) RatesResult {
    const x = s.pos.x;
    const y = s.pos.y;
    const vx = s.vel.x;
    const vy = s.vel.y;
    const m = s.mass;
    const r1_3, const r2_3 = gravity(x, y);
    const v = math.sqrt(vx * vx + vy * vy);
    const tmv = th.thrust / (m * v);
    const ax = ((2 * W * vy + W * W * x) - mu1 * (x - x1) / r1_3) - mu2 * (x - x2) / r2_3 + tmv * vx;
    const ay = (-2 * W * vx + W * W * y) - (mu1 / r1_3 + mu2 / r2_3) * y + tmv * vy;
    return RatesResult{
        .d_pos = s.vel,
        .d_vel = Vec{ .x = ax, .y = ay },
        .dm = th.massRate(),
    };
}

fn ratesCoastEM(s: State) RatesResult {
    const x = s.pos.x;
    const y = s.pos.y;
    const vx = s.vel.x;
    const vy = s.vel.y;
    const r1_3, const r2_3 = gravity(x, y);
    const ax = ((2 * W * vy + W * W * x) - mu1 * (x - x1) / r1_3) - mu2 * (x - x2) / r2_3;
    const ay = (-2 * W * vx + W * W * y) - (mu1 / r1_3 + mu2 / r2_3) * y;
    return RatesResult{
        .d_pos = s.vel,
        .d_vel = Vec{ .x = ax, .y = ay },
        .dm = 0,
    };
}

fn ratesBrakeEM(s: State, th: Thruster) RatesResult {
    const x = s.pos.x;
    const y = s.pos.y;
    const vx = s.vel.x;
    const vy = s.vel.y;
    const m = s.mass;
    const r1_3, const r2_3 = gravity(x, y);
    const v = math.sqrt(vx * vx + vy * vy);
    const tmv = -th.thrust / (m * v); // negative: retrograde
    const ax = ((2 * W * vy + W * W * x) - mu1 * (x - x1) / r1_3) - mu2 * (x - x2) / r2_3 + tmv * vx;
    const ay = (-2 * W * vx + W * W * y) - (mu1 / r1_3 + mu2 / r2_3) * y + tmv * vy;
    return RatesResult{
        .d_pos = s.vel,
        .d_vel = Vec{ .x = ax, .y = ay },
        .dm = th.massRate(),
    };
}

// EventFunc equivalent: union captures closure data.
const Event = union(enum) {
    jacobi_em: f64, // threshold
    l1_em: void,

    fn eval(ev: Event, t: f64, s: State, phi_s0: f64) f64 {
        _ = t;
        _ = phi_s0;
        return switch (ev) {
            .jacobi_em => |thr| jacobiConstantEM(s) - thr,
            .l1_em => math.hypot(s.pos.x - x1, s.pos.y) - l1dist,
        };
    }
};

fn jacobiConstantEM(s: State) f64 {
    const v2 = v3Norm2(s.vel);
    const d1 = math.hypot(s.pos.x - x1, s.pos.y);
    const d2 = math.hypot(s.pos.x - x2, s.pos.y);
    return 0.5 * v2 - 0.5 * W * W * (s.pos.x * s.pos.x + s.pos.y * s.pos.y) - mu1 / d1 - mu2 / d2;
}

fn gravity(x: f64, y: f64) struct { f64, f64 } {
    const dx1 = x - x1;
    const dx2 = x - x2;
    const r1 = math.sqrt(dx1 * dx1 + y * y);
    const r2 = math.sqrt(dx2 * dx2 + y * y);
    return .{ r1 * r1 * r1, r2 * r2 * r2 };
}

// Dormand-Prince RK45 coefficients.
const dp_c = [7]f64{ 0, 1.0 / 5.0, 3.0 / 10.0, 4.0 / 5.0, 8.0 / 9.0, 1, 1 };

const dp_a = [7][6]f64{
    .{ 0, 0, 0, 0, 0, 0 },
    .{ 1.0 / 5.0, 0, 0, 0, 0, 0 },
    .{ 3.0 / 40.0, 9.0 / 40.0, 0, 0, 0, 0 },
    .{ 44.0 / 45.0, -56.0 / 15.0, 32.0 / 9.0, 0, 0, 0 },
    .{ 19372.0 / 6561.0, -25360.0 / 2187.0, 64448.0 / 6561.0, -212.0 / 729.0, 0, 0 },
    .{ 9017.0 / 3168.0, -355.0 / 33.0, 46732.0 / 5247.0, 49.0 / 176.0, -5103.0 / 18656.0, 0 },
    .{ 35.0 / 384.0, 0, 500.0 / 1113.0, 125.0 / 192.0, -2187.0 / 6784.0, 11.0 / 84.0 },
};

const dp_b = [7]f64{ 35.0 / 384.0, 0, 500.0 / 1113.0, 125.0 / 192.0, -2187.0 / 6784.0, 11.0 / 84.0, 0 };

const dp_e = [7]f64{
    71.0 / 57600.0,
    0,
    -71.0 / 16695.0,
    71.0 / 1920.0,
    -17253.0 / 339200.0,
    22.0 / 525.0,
    -1.0 / 40.0,
};

const Integrator = struct {
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

    fn step(ig: *Integrator, h_in: f64) f64 {
        const safety = 0.9;
        const min_fac = 0.2;
        const max_fac = 10.0;
        const order = 5.0;

        const t = ig.t;
        const s = ig.state;
        const phi_s0 = ig.phi_s0;

        var h = h_in;
        var step_rejected = false;
        while (true) {
            const t_next = t + h;
            const h_eff = t_next - t;

            var k_pos: [7]Vec = undefined;
            var k_vel: [7]Vec = undefined;
            var k_m: [7]f64 = undefined;

            const r0 = ig.rates.compute(t, s, phi_s0);
            k_pos[0] = r0.d_pos;
            k_vel[0] = r0.d_vel;
            k_m[0] = r0.dm;

            for (1..7) |i| {
                var sum_pos = Vec{};
                var sum_vel = Vec{};
                var sum_mass: f64 = 0;
                for (0..i) |j| {
                    sum_pos.x += dp_a[i][j] * k_pos[j].x;
                    sum_pos.y += dp_a[i][j] * k_pos[j].y;
                    sum_vel.x += dp_a[i][j] * k_vel[j].x;
                    sum_vel.y += dp_a[i][j] * k_vel[j].y;
                    sum_mass += dp_a[i][j] * k_m[j];
                }
                const ti = t + dp_c[i] * h_eff;
                const si = State{
                    .pos = v3Add(s.pos, v3Scale(h_eff, sum_pos)),
                    .vel = v3Add(s.vel, v3Scale(h_eff, sum_vel)),
                    .mass = s.mass + h_eff * sum_mass,
                };
                const ri = ig.rates.compute(ti, si, phi_s0);
                k_pos[i] = ri.d_pos;
                k_vel[i] = ri.d_vel;
                k_m[i] = ri.dm;
            }

            var sum_b_pos = Vec{};
            var sum_b_vel = Vec{};
            var sum_e_pos = Vec{};
            var sum_e_vel = Vec{};
            var sum_bm: f64 = 0;
            var sum_em: f64 = 0;
            for (0..7) |i| {
                sum_b_pos.x += dp_b[i] * k_pos[i].x;
                sum_b_pos.y += dp_b[i] * k_pos[i].y;
                sum_b_vel.x += dp_b[i] * k_vel[i].x;
                sum_b_vel.y += dp_b[i] * k_vel[i].y;
                sum_bm += dp_b[i] * k_m[i];
                sum_e_pos.x += dp_e[i] * k_pos[i].x;
                sum_e_pos.y += dp_e[i] * k_pos[i].y;
                sum_e_vel.x += dp_e[i] * k_vel[i].x;
                sum_e_vel.y += dp_e[i] * k_vel[i].y;
                sum_em += dp_e[i] * k_m[i];
            }

            const new_pos = v3Add(s.pos, v3Scale(h_eff, sum_b_pos));
            const new_vel = v3Add(s.vel, v3Scale(h_eff, sum_b_vel));
            const new_m = s.mass + h_eff * sum_bm;
            const err_pos = v3Scale(h_eff, sum_e_pos);
            const err_vel = v3Scale(h_eff, sum_e_vel);
            const err_m = h_eff * sum_em;

            const sc_px = ig.atol + ig.rtol * @max(@abs(s.pos.x), @abs(new_pos.x));
            const sc_py = ig.atol + ig.rtol * @max(@abs(s.pos.y), @abs(new_pos.y));
            const sc_vx = ig.atol + ig.rtol * @max(@abs(s.vel.x), @abs(new_vel.x));
            const sc_vy = ig.atol + ig.rtol * @max(@abs(s.vel.y), @abs(new_vel.y));
            const sc_m = ig.atol + ig.rtol * @max(@abs(s.mass), @abs(new_m));

            const sum_sq = (err_pos.x / sc_px) * (err_pos.x / sc_px) +
                (err_pos.y / sc_py) * (err_pos.y / sc_py) +
                (err_vel.x / sc_vx) * (err_vel.x / sc_vx) +
                (err_vel.y / sc_vy) * (err_vel.y / sc_vy) +
                (err_m / sc_m) * (err_m / sc_m);
            const err_norm = math.sqrt(sum_sq) / math.sqrt(5.0);

            if (err_norm <= 1) {
                ig.t = t_next;
                ig.state = State{ .pos = new_pos, .vel = new_vel, .mass = new_m };
                ig.last_err_norm = err_norm;
                ig.step_count += 1;

                if (err_norm == 0) {
                    return @min(h_eff * max_fac, ig.max_step);
                }
                const factor_raw = safety * math.pow(f64, err_norm, -1.0 / order);
                var factor_clamped = @max(min_fac, @min(max_fac, factor_raw));
                if (step_rejected) factor_clamped = @min(1.0, factor_clamped);
                return @max(ig.min_step, @min(ig.max_step, h_eff * factor_clamped));
            }

            // Step rejected.
            var factor = safety * math.pow(f64, err_norm, -1.0 / order);
            factor = @max(min_fac, factor);
            h = @max(ig.min_step, h_eff * factor);
            step_rejected = true;

            if (h <= ig.min_step) {
                ig.t = t_next;
                ig.state = State{ .pos = new_pos, .vel = new_vel, .mass = new_m };
                return ig.min_step;
            }
        }
    }
};

// IntegrateUntil result mirrors Go's (eventIdx int, eventTime float64).
const IntegrateResult = struct {
    event_idx: i32,
    event_time: f64,
};

fn integrateUntil(ig: *Integrator, tf: f64, events: []const Event) IntegrateResult {
    var h = ig.max_step;

    var prev_events: [4]f64 = undefined;
    for (events, 0..) |ev, i| {
        prev_events[i] = ev.eval(ig.t, ig.state, ig.phi_s0);
    }

    while (ig.t < tf) {
        if (ig.t + h > tf) {
            h = tf - ig.t;
        }
        const t_prev = ig.t;
        const s_prev = ig.state;
        h = ig.step(h);

        for (events, 0..) |ev, i| {
            const curr = ev.eval(ig.t, ig.state, ig.phi_s0);
            if (prev_events[i] * curr < 0) {
                const event_time = bisectEvent(t_prev, ig.t, s_prev, ig.state, ig.phi_s0, ev);
                const alpha = (event_time - t_prev) / (ig.t - t_prev);
                ig.state = State{
                    .pos = v3Add(v3Scale(1 - alpha, s_prev.pos), v3Scale(alpha, ig.state.pos)),
                    .vel = v3Add(v3Scale(1 - alpha, s_prev.vel), v3Scale(alpha, ig.state.vel)),
                    .mass = (1 - alpha) * s_prev.mass + alpha * ig.state.mass,
                };
                ig.t = event_time;
                return IntegrateResult{ .event_idx = @intCast(i), .event_time = event_time };
            }
            prev_events[i] = curr;
        }
    }

    return IntegrateResult{ .event_idx = -1, .event_time = ig.t };
}

fn bisectEvent(t0_in: f64, t1_in: f64, s0_in: State, s1_in: State, phi_s0: f64, ev: Event) f64 {
    const max_iter = 50;
    const tol = 1e-10;

    var t0 = t0_in;
    var t1 = t1_in;
    var s0 = s0_in;
    var s1 = s1_in;

    for (0..max_iter) |_| {
        const tm = 0.5 * (t0 + t1);
        if (t1 - t0 < tol) {
            return tm;
        }
        const alpha = (tm - t0) / (t1 - t0);
        const sm = State{
            .pos = v3Add(v3Scale(1 - alpha, s0.pos), v3Scale(alpha, s1.pos)),
            .vel = v3Add(v3Scale(1 - alpha, s0.vel), v3Scale(alpha, s1.vel)),
            .mass = (1 - alpha) * s0.mass + alpha * s1.mass,
        };
        const vm = ev.eval(tm, sm, phi_s0);
        const v0 = ev.eval(t0, s0, phi_s0);
        if (v0 * vm < 0) {
            t1 = tm;
            s1 = sm;
        } else {
            t0 = tm;
            s0 = sm;
        }
    }
    return 0.5 * (t0 + t1);
}

const GTOConfig = struct {
    h_apogee: f64,
    h_perigee: f64,
    phi: f64,
    gamma: f64,
    m0: f64,
    thruster: Thruster,
    jacobi_thr: f64,
    tol: f64,

    fn gtoInitialState(cfg: *const GTOConfig) State {
        const r_apogee = r_earth + cfg.h_apogee;
        const r_perigee = r_earth + cfg.h_perigee;
        const e = (r_apogee - r_perigee) / (r_apogee + r_perigee);
        const v0 = math.sqrt(mu1 * (1 - e) / r_apogee) - W * r_apogee;

        const sin_phi = @sin(cfg.phi);
        const cos_phi = @cos(cfg.phi);
        const sin_gam = @sin(cfg.gamma);
        const cos_gam = @cos(cfg.gamma);

        return State{
            .pos = Vec{ .x = r_apogee * cos_phi + x1, .y = r_apogee * sin_phi },
            .vel = Vec{
                .x = v0 * (sin_gam * cos_phi - cos_gam * sin_phi),
                .y = v0 * (sin_gam * sin_phi + cos_gam * cos_phi),
            },
            .mass = cfg.m0,
        };
    }
};

fn printState(phase: i32, t: f64, s: State) void {
    stdout.print("phase={d} t={d:6.2}d earthdist={d:.1}km pos=({d:.3},{d:.3})\n", .{
        phase, t / days, v3Norm(s.pos), s.pos.x, s.pos.y,
    });
}

fn run(max_step_arg: f64, verify: bool) !void {
    const cfg = GTOConfig{
        .h_apogee = 37000,
        .h_perigee = 1200,
        .phi = 0.7505211952744961 * std.math.pi / 180.0,
        .gamma = 0,
        .m0 = 12,
        .thruster = Thruster{ .thrust = 4 * 0.000000450, .isp = 1650 },
        .jacobi_thr = -1.63907788,
        .tol = 1e-12,
    };

    var ig = Integrator{
        .t = 0,
        .state = cfg.gtoInitialState(),
        .phi_s0 = 0,
        .rates = RatesMode{ .thrust = cfg.thruster },
        .min_step = 0,
        .max_step = @min(450, max_step_arg),
        .atol = cfg.tol,
        .rtol = 1e-9,
        .last_err_norm = 0,
        .step_count = 0,
    };

    if (verify) printState(0, 0, ig.state);

    // Phase 1: prograde thrust until Jacobi threshold.
    const ev1 = [_]Event{.{ .jacobi_em = cfg.jacobi_thr }};
    const r1 = integrateUntil(&ig, days * 360, &ev1);
    if (r1.event_idx < 0) return error.Phase1Failed;
    const t1 = r1.event_time;
    if (verify) printState(1, t1, ig.state);

    // Phase 2: coast until distance from earth = L1.
    ig.rates = RatesMode{ .coast = {} };
    ig.max_step = @min(200, max_step_arg);
    const ev2 = [_]Event{.{ .l1_em = {} }};
    const r2 = integrateUntil(&ig, t1 + 260 * days, &ev2);
    if (r2.event_idx < 0) return error.Phase2Failed;
    const t2 = r2.event_time;
    if (verify) printState(2, t2, ig.state);

    // Phase 3: retrograde braking until Jacobi constant reaches C1.
    ig.rates = RatesMode{ .brake = cfg.thruster };
    ig.max_step = @min(1 * days, max_step_arg);
    const ev3 = [_]Event{.{ .jacobi_em = c1 }};
    const r3 = integrateUntil(&ig, t2 + 25 * days, &ev3);
    if (r3.event_idx < 0) return error.Phase3Failed;
    const t3 = r3.event_time;
    if (verify) printState(3, t3, ig.state);

    // Phase 4: 20-day coast, no events.
    ig.rates = RatesMode{ .coast = {} };
    ig.max_step = @min(100, max_step_arg);
    const r4 = integrateUntil(&ig, t3 + 20 * days, &[_]Event{});
    const t4 = r4.event_time;
    if (verify) {
        printState(4, t4, ig.state);
        stdout.print("num steps={d}\n", .{ig.step_count});
    }
}

pub fn main(init: std.process.Init) !void {
    const allocator = init.gpa;

    var it = try init.minimal.args.iterateAllocator(allocator);
    defer it.deinit();

    _ = it.skip(); // program name

    const arg = it.next() orelse return error.MissingArgs;
    const max_step = try std.fmt.parseFloat(f64, arg);
    const verify_arg = it.next();
    const verify = verify_arg != null and std.mem.eql(u8, verify_arg.?, "v");

    _ = n_s; // mirrors Go's var nS (computed but unused)

    run(max_step, verify) catch |err| {
        switch (err) {
            error.Phase1Failed => stdout.print("phase 1: jacobi threshold not crossed\n", .{}),
            error.Phase2Failed => stdout.print("during phase 2: L1 distance not reached\n", .{}),
            error.Phase3Failed => stdout.print("during phase 3: C1 distance not reached while braking\n", .{}),
            else => |e| return e,
        }
        std.process.exit(1);
    };
}

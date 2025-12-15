const std = @import("std");
const math = std.math;
const stdout = std.debug;

var gpa = std.heap.GeneralPurposeAllocator(.{}){};
const allocator = gpa.allocator();

const pi = 3.141592653589793;
const solarMass = 4 * pi * pi;
const daysPerYear = 365.24;
const tol = 1e-5;

const Planet = struct {
    x: f64,
    y: f64,
    z: f64,
    vx: f64,
    vy: f64,
    vz: f64,
    mass: f64,
};

fn sqrt_newton(v: f64) f64 {
    if (v < 0) {
        return -1.0;
    }
    var x: f64 = v;
    while (true) {
        const delta = (x * x - v) / (2 * x);
        x -= delta;

        if (@abs(delta) <= tol) {
            break;
        }
    }
    return x;
}

fn advance(bodies: []Planet, dt: f64) void {
    for (0..bodies.len) |i| {
        var b = &bodies[i];
        for (i + 1..bodies.len) |j| {
            var b2 = &bodies[j];
            const dx = b.x - b2.x;
            const dy = b.y - b2.y;
            const dz = b.z - b2.z;

            const distance_sq = dx * dx + dy * dy + dz * dz;
            const distance = sqrt_newton(distance_sq);
            const mag = dt / (distance_sq * distance);

            b.vx -= dx * b2.mass * mag;
            b.vy -= dy * b2.mass * mag;
            b.vz -= dz * b2.mass * mag;

            b2.vx += dx * b.mass * mag;
            b2.vy += dy * b.mass * mag;
            b2.vz += dz * b.mass * mag;
        }
    }

    for (bodies) |*b| {
        b.x += dt * b.vx;
        b.y += dt * b.vy;
        b.z += dt * b.vz;
    }
}

fn energy(bodies: []Planet) f64 {
    var e: f64 = 0.0;
    for (0..bodies.len) |i| {
        const b = bodies[i];
        e += 0.5 * b.mass * (b.vx * b.vx + b.vy * b.vy + b.vz * b.vz);

        for (i + 1..bodies.len) |j| {
            const b2 = bodies[j];
            const dx = b.x - b2.x;
            const dy = b.y - b2.y;
            const dz = b.z - b2.z;
            const distance = sqrt_newton(dx * dx + dy * dy + dz * dz);
            e -= (b.mass * b2.mass) / distance;
        }
    }
    return e;
}

fn offsetMomentum(bodies: []Planet) void {
    var px: f64 = 0.0;
    var py: f64 = 0.0;
    var pz: f64 = 0.0;

    for (bodies) |b| {
        px += b.vx * b.mass;
        py += b.vy * b.mass;
        pz += b.vz * b.mass;
    }

    bodies[0].vx = -px / solarMass;
    bodies[0].vy = -py / solarMass;
    bodies[0].vz = -pz / solarMass;
}

var Bodies = [_]Planet{
    .{ // Sun
        .x = 0.0,
        .y = 0.0,
        .z = 0.0,
        .vx = 0.0,
        .vy = 0.0,
        .vz = 0.0,
        .mass = solarMass,
    },
    .{ // Jupiter
        .x = 4.84143144246472090e+00,
        .y = -1.16032004402742839e+00,
        .z = -1.03622044471123109e-01,
        .vx = 1.66007664274403694e-03 * daysPerYear,
        .vy = 7.69901118419740425e-03 * daysPerYear,
        .vz = -6.90460016972063023e-05 * daysPerYear,
        .mass = 9.54791938424326609e-04 * solarMass,
    },
    .{ // Saturn
        .x = 8.34336671824457987e+00,
        .y = 4.12479856412430479e+00,
        .z = -4.03523417114321381e-01,
        .vx = -2.76742510726862411e-03 * daysPerYear,
        .vy = 4.99852801234917238e-03 * daysPerYear,
        .vz = 2.30417297573763929e-05 * daysPerYear,
        .mass = 2.85885980666130812e-04 * solarMass,
    },
    .{ // Uranus
        .x = 1.28943695621391310e+01,
        .y = -1.51111514016986312e+01,
        .z = -2.23307578892655734e-01,
        .vx = 2.96460137564761618e-03 * daysPerYear,
        .vy = 2.37847173959480950e-03 * daysPerYear,
        .vz = -2.96589568540237556e-05 * daysPerYear,
        .mass = 4.36624404335156298e-05 * solarMass,
    },
    .{ // Neptune
        .x = 1.53796971148509165e+01,
        .y = -2.59193146099879641e+01,
        .z = 1.79258772950371181e-01,
        .vx = 2.68067772490389322e-03 * daysPerYear,
        .vy = 1.62824170038242295e-03 * daysPerYear,
        .vz = -9.51592254519715870e-05 * daysPerYear,
        .mass = 5.15138902046611451e-05 * solarMass,
    },
};

pub fn main() !void {
    const args = try std.process.argsAlloc(allocator);

    if (args.len < 2) {
        std.debug.print("Usage: {s} <iterations>\n", .{args[0]});
        return error.InvalidArgs;
    }

    const n = try std.fmt.parseInt(usize, args[1], 10);
    const bodies = Bodies[0..];

    offsetMomentum(bodies);
    stdout.print("{d:.9}\n", .{energy(bodies)});

    for (0..n) |_| {
        advance(bodies, 0.01);
    }

    stdout.print("{d:.9}\n", .{energy(bodies)});
}

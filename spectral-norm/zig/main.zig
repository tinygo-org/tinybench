const std = @import("std");
const math = std.math;
const stdout = std.debug;

fn evala(i: usize, j: usize) usize {
    return (i + j) * (i + j + 1) / 2 + i + 1;
}

fn times(v: []f64, u: []const f64) void {
    for (v, 0..) |*vi, i| {
        var a: f64 = 0.0;
        for (u, 0..) |uj, j| {
            a += uj / @as(f64, @floatFromInt(evala(i, j)));
        }
        vi.* = a;
    }
}

fn timesTrans(v: []f64, u: []const f64) void {
    for (v, 0..) |*vi, i| {
        var a: f64 = 0.0;
        for (u, 0..) |uj, j| {
            a += uj / @as(f64, @floatFromInt(evala(j, i)));
        }
        vi.* = a;
    }
}

fn aTimesTransp(allocator: std.mem.Allocator, v: []f64, u: []const f64) !void {
    const x = try allocator.alloc(f64, u.len);
    defer allocator.free(x);
    times(x, u);
    timesTrans(v, x);
}

pub fn main(init: std.process.Init) !void {
    const allocator = init.gpa;

    var it = try init.minimal.args.iterateAllocator(allocator);
    defer it.deinit();

    _ = it.skip(); // program name

    const arg = it.next() orelse return error.MissingArgs;
    const n = try std.fmt.parseInt(usize, arg, 10);

    const u = try allocator.alloc(f64, n);
    const v = try allocator.alloc(f64, n);
    defer allocator.free(u);
    defer allocator.free(v);

    for (u) |*ui| ui.* = 1.0;
    for (v) |*vi| vi.* = 1.0;

    for (0..10) |_| {
        try aTimesTransp(allocator, v, u);
        try aTimesTransp(allocator, u, v);
    }

    var vBv: f64 = 0.0;
    var vv: f64 = 0.0;
    for (v, 0..) |vi, i| {
        vBv += u[i] * vi;
        vv += vi * vi;
    }

    stdout.print("{d:.9}\n", .{math.sqrt(vBv / vv)});
}

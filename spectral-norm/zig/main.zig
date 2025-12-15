const std = @import("std");
const math = std.math;
const stdout = std.debug;

var gpa = std.heap.GeneralPurposeAllocator(.{}){};
const allocator = gpa.allocator();

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

fn aTimesTransp(v: []f64, u: []const f64) !void {
    const x = try allocator.alloc(f64, u.len);
    defer allocator.free(x);
    times(x, u);
    timesTrans(v, x);
}

pub fn main() !void {
    const args = try std.process.argsAlloc(allocator);

    if (args.len < 2) {
        std.debug.print("Usage: {s} <n>\n", .{args[0]});
        return error.InvalidArgs;
    }

    const n = try std.fmt.parseInt(usize, args[1], 10);

    const u = try allocator.alloc(f64, n);
    const v = try allocator.alloc(f64, n);
    defer allocator.free(u);
    defer allocator.free(v);

    for (u) |*ui| ui.* = 1.0;
    for (v) |*vi| vi.* = 1.0;

    for (0..10) |_| {
        try aTimesTransp(v, u);
        try aTimesTransp(u, v);
    }

    var vBv: f64 = 0.0;
    var vv: f64 = 0.0;
    for (v, 0..) |vi, i| {
        vBv += u[i] * vi;
        vv += vi * vi;
    }

    stdout.print("{d:.9}\n", .{math.sqrt(vBv / vv)});
}

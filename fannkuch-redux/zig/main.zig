const std = @import("std");
const stdout = std.debug;
const stderr = std.debug;

var gpa = std.heap.GeneralPurposeAllocator(.{}){};
const allocator = gpa.allocator();

const Fannkuchen = struct {
    s: [16]i64,
    t: [16]i64,
    maxflips: i64,
    max_n: usize,
    odd: i64,
    checksum: i64,

    fn flip(self: *Fannkuchen) i64 {
        @memcpy(self.t[0..self.max_n], self.s[0..self.max_n]);
        var flips: i64 = 1;

        while (true) {
            const y = @as(usize, @intCast(self.t[0]));
            var x: usize = 0;
            var y_idx = y;

            // Reverse elements
            while (x < y_idx) {
                const tmp = self.t[x];
                self.t[x] = self.t[y_idx];
                self.t[y_idx] = tmp;
                x += 1;
                y_idx -= 1;
            }

            flips += 1;
            const check_idx = @as(usize, @intCast(self.t[0]));
            if (self.t[check_idx] == 0) break;
        }
        return flips;
    }

    fn rotate(self: *Fannkuchen, n: usize) void {
        const c = self.s[0];
        for (1..n + 1) |i| {
            self.s[i - 1] = self.s[i];
        }
        self.s[n] = c;
    }

    fn tk(self: *Fannkuchen) void {
        var c: [16]i64 = [_]i64{0} ** 16;
        var i: usize = 0;

        while (i < self.max_n) {
            self.rotate(i);

            if (c[i] >= @as(i64, @intCast(i))) {
                c[i] = 0;
                i += 1;
                continue;
            }

            c[i] += 1;
            i = 1;
            self.odd = ~self.odd;

            if (self.s[0] != 0) {
                const s0 = @as(usize, @intCast(self.s[0]));
                const f = if (self.s[s0] != 0) self.flip() else 1;

                if (f > self.maxflips) self.maxflips = f;
                self.checksum += if (self.odd != 0) -f else f;
            }
        }
    }
};

pub fn main() !void {
    const args = try std.process.argsAlloc(allocator);
    defer std.process.argsFree(allocator, args);

    if (args.len < 2) {
        stderr.print("Usage: {s} <n>\n", .{args[0]});
        return error.InvalidArgs;
    }

    const max_n = std.fmt.parseInt(usize, args[1], 10) catch {
        stderr.print("Invalid number: {s}\n", .{args[1]});
        return error.InvalidNumber;
    };

    if (max_n < 1 or max_n > 15) {
        stderr.print("n must be 1-15\n", .{});
        return error.InvalidRange;
    }

    var f = Fannkuchen{
        .s = undefined,
        .t = undefined,
        .maxflips = 0,
        .max_n = max_n,
        .odd = 0,
        .checksum = 0,
    };

    // Initialize permutation
    for (0..max_n) |i| {
        f.s[i] = @intCast(i);
    }

    f.tk();

    stdout.print("{d}\nPfannkuchen({d}) = {d}\n", .{ f.checksum, max_n, f.maxflips });
}

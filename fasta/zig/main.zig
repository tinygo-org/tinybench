const std = @import("std");
const stdout = std.debug;

const WIDTH = 60;

const AminoAcid = struct {
    p: f64,
    c: u8,
};

fn accumulateProbabilities(genelist: []AminoAcid) void {
    for (genelist[1..], 0..) |*aa, i| {
        aa.p += genelist[i].p;
    }
}

fn repeatFasta(allocator: std.mem.Allocator, s: []const u8, count: usize, out: anytype, verify: bool) !void {
    var pos: usize = 0;
    var s2 = try allocator.alloc(u8, s.len + WIDTH);
    defer allocator.free(s2);

    @memcpy(s2[0..s.len], s);
    @memcpy(s2[s.len .. s.len + WIDTH], s[0..WIDTH]);

    var remaining = count;
    while (remaining > 0) {
        const line = @min(WIDTH, remaining);
        if (verify) {
            out.print("{s}\n", .{s2[pos .. pos + line]});
        }
        pos += line;
        if (pos >= s.len) pos -= s.len;
        remaining -= line;
    }
}

var lastrandom: u32 = 42;
const IM: u32 = 139968;
const IA: u32 = 3877;
const IC: u32 = 29573;

fn randomFasta(genelist: []const AminoAcid, count: usize, out: anytype, verify: bool) !void {
    var buf: [WIDTH + 1]u8 = undefined;
    var remaining = count;
    while (remaining > 0) {
        const line = @min(WIDTH, remaining);
        for (buf[0..line]) |*b| {
            lastrandom = (lastrandom * IA + IC) % IM;
            const r = @as(f64, @floatFromInt(lastrandom)) / @as(f64, IM);
            for (genelist) |aa| {
                if (aa.p >= r) {
                    b.* = aa.c;
                    break;
                }
            }
        }
        buf[line] = '\n';
        if (verify) {
            out.print("{s}", .{buf[0 .. line + 1]});
        }
        remaining -= line;
    }
}

pub fn main(init: std.process.Init) !void {
    const allocator = init.gpa;

    var it = try init.minimal.args.iterateAllocator(allocator);
    defer it.deinit();

    _ = it.skip(); // program name

    const arg = it.next() orelse return error.MissingArgs;
    const max_n = try std.fmt.parseInt(usize, arg, 10);
    if (max_n < 1 or max_n > 50000000) {
        return error.InvalidRange;
    }

    const verify_arg = it.next();
    const verify = verify_arg != null and std.mem.eql(u8, verify_arg.?, "v");

    var iub = [_]AminoAcid{
        AminoAcid{ .p = 0.27, .c = 'a' },
        AminoAcid{ .p = 0.12, .c = 'c' },
        AminoAcid{ .p = 0.12, .c = 'g' },
        AminoAcid{ .p = 0.27, .c = 't' },
        AminoAcid{ .p = 0.02, .c = 'B' },
        AminoAcid{ .p = 0.02, .c = 'D' },
        AminoAcid{ .p = 0.02, .c = 'H' },
        AminoAcid{ .p = 0.02, .c = 'K' },
        AminoAcid{ .p = 0.02, .c = 'M' },
        AminoAcid{ .p = 0.02, .c = 'N' },
        AminoAcid{ .p = 0.02, .c = 'R' },
        AminoAcid{ .p = 0.02, .c = 'S' },
        AminoAcid{ .p = 0.02, .c = 'V' },
        AminoAcid{ .p = 0.02, .c = 'W' },
        AminoAcid{ .p = 0.02, .c = 'Y' },
    };

    var homosapiens = [_]AminoAcid{
        AminoAcid{ .p = 0.3029549426680, .c = 'a' },
        AminoAcid{ .p = 0.1979883004921, .c = 'c' },
        AminoAcid{ .p = 0.1975473066391, .c = 'g' },
        AminoAcid{ .p = 0.3015094502008, .c = 't' },
    };

    accumulateProbabilities(iub[0..]);
    accumulateProbabilities(homosapiens[0..]);

    const alu = "GGCCGGGCGCGGTGGCTCACGCCTGTAATCCCAGCACTTTGGGAGGCCGAGGCGGGCGGATCACCTGAGGTC" ++
        "AGGAGTTCGAGACCAGCCTGGCCAACATGGTGAAACCCCGTCTCTACTAAAAATACAAAAATTAGCCGGGCG" ++
        "TGGTGGCGCGCGCCTGTAATCCCAGCTACTCGGGAGGCTGAGGCAGGAGAATCGCTTGAACCCGGGAGGCGG" ++
        "AGGTTGCAGTGAGCCGAGATCGCGCCACTGCACTCCAGCCTGGGCGACAGAGCGAGACTCCGTCTCAAAAA";

    if (verify) stdout.print(">ONE Homo sapiens alu\n", .{});
    try repeatFasta(allocator, alu, 2 * max_n, stdout, verify);

    if (verify) stdout.print(">TWO IUB ambiguity codes\n", .{});
    try randomFasta(iub[0..], 3 * max_n, stdout, verify);

    if (verify) stdout.print(">THREE Homo sapiens frequency\n", .{});
    try randomFasta(homosapiens[0..], 5 * max_n, stdout, verify);
}

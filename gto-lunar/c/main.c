// Original Author: Juan Pablo Sellanes Goncalves
// https://eventos.iua.edu.ar/event/1/contributions/51/

#include <math.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#ifndef M_PI
#define M_PI 3.14159265358979323846
#endif

/* Physical constants */
#define g0      9.807e-3
#define days    86400.0
#define bigG    6.6742e-20
#define r12     384400.0
#define m1      5.974e24
#define m2      7.348e22
#define mu1     398600.0
#define mu2     4903.02
#define mS      1.989e30
#define rB2S    149597870.7
#define rMoon   1737.0
#define rEarth  6378.0
#define L1dist  321710.0
#define C1      (-1.676)

/* Runtime-derived globals (require sqrt or floating-point arithmetic) */
static double W;   /* [rad/s] angular velocity of rotating frame */
static double nS;  /* [rad/s] Sun mean motion */
static double x1;  /* [km] Earth x-position in rotating frame */
static double x2;  /* [km] Moon x-position in rotating frame */

/* --- Data types matching Go struct layout --- */

typedef struct { double X, Y, Z; } Vec;

typedef struct {
    Vec    Pos;
    Vec    Vel;
    double Mass;
} State;

typedef struct {
    double Thrust;
    double Isp;
} Thruster;

typedef struct {
    double   HApogee, HPerigee;
    double   Phi, Gamma;
    double   M0;
    Thruster Thruster;
    double   JacobiThr;
    double   Tol;
} GTOConfig;

/*
 * RatesFunc: C equivalent of Go's closure-based RatesFunc.
 * ctx captures closure state (e.g. Thruster*); NULL for coast.
 */
typedef void (*RatesFunc)(double t, State s, double phiS0,
                          Vec *dPos, Vec *dVel, double *dm, void *ctx);

/* EventFunc: returns a value that crosses zero at an event. */
typedef double (*EventFunc)(double t, State s, double phiS0, void *ctx);

typedef struct {
    EventFunc fn;
    void     *ctx;
} Event;

typedef struct {
    double    T;
    State     State;
    double    PhiS0;
    RatesFunc Rates;
    void     *RatesCtx;
    double    MinStep, MaxStep;
    double    ATol, RTol;
    double    LastErrNorm;
    int       StepCount;
} Integrator;

/* --- Thruster --- */

static double MassRate(Thruster th) { return -th.Thrust / (g0 * th.Isp); }

/* --- Vector operations --- */

static double v3Norm2(Vec v) { return v.X*v.X + v.Y*v.Y + v.Z*v.Z; }
static double v3Norm(Vec v)  { return hypot(v.X, hypot(v.Y, v.Z)); }

static Vec v3Sub(Vec a, Vec b) {
    return (Vec){ a.X-b.X, a.Y-b.Y, a.Z-b.Z };
}
static Vec v3Add(Vec a, Vec b) {
    return (Vec){ a.X+b.X, a.Y+b.Y, a.Z+b.Z };
}
static Vec v3Scale(double f, Vec v) {
    return (Vec){ f*v.X, f*v.Y, f*v.Z };
}
static Vec v3FMA(double a, Vec b, Vec c) {
    return (Vec){ fma(a, b.X, c.X), fma(a, b.Y, c.Y), fma(a, b.Z, c.Z) };
}

/* --- gravity --- */

static void gravity(double gx, double y, double *r1_3, double *r2_3) {
    double dx1 = gx - x1;
    double dx2 = gx - x2;
    double r1 = sqrt(dx1*dx1 + y*y);
    double r2 = sqrt(dx2*dx2 + y*y);
    *r1_3 = r1 * r1 * r1;
    *r2_3 = r2 * r2 * r2;
}

/* --- Rates functions --- */

static void RatesThrustEM(double t, State s, double phiS0,
                          Vec *dPos, Vec *dVel, double *dm, void *ctx) {
    Thruster *th = (Thruster *)ctx;
    double px = s.Pos.X, py = s.Pos.Y;
    double vx = s.Vel.X, vy = s.Vel.Y, mass = s.Mass;
    double r1_3, r2_3;
    gravity(px, py, &r1_3, &r2_3);
    double v = sqrt(vx*vx + vy*vy);
    double tmv = th->Thrust / (mass * v);
    double ax = ((2*W*vy + W*W*px) - mu1*(px-x1)/r1_3) - mu2*(px-x2)/r2_3 + tmv*vx;
    double ay = (-2*W*vx + W*W*py) - (mu1/r1_3+mu2/r2_3)*py + tmv*vy;
    *dPos = s.Vel;
    *dVel = (Vec){ ax, ay, 0 };
    *dm   = MassRate(*th);
}

static void RatesCoastEM(double t, State s, double phiS0,
                         Vec *dPos, Vec *dVel, double *dm, void *ctx) {
    double px = s.Pos.X, py = s.Pos.Y;
    double vx = s.Vel.X, vy = s.Vel.Y;
    double r1_3, r2_3;
    gravity(px, py, &r1_3, &r2_3);
    double ax = ((2*W*vy + W*W*px) - mu1*(px-x1)/r1_3) - mu2*(px-x2)/r2_3;
    double ay = (-2*W*vx + W*W*py) - (mu1/r1_3+mu2/r2_3)*py;
    *dPos = s.Vel;
    *dVel = (Vec){ ax, ay, 0 };
    *dm   = 0;
}

static void RatesBrakeEM(double t, State s, double phiS0,
                         Vec *dPos, Vec *dVel, double *dm, void *ctx) {
    Thruster *th = (Thruster *)ctx;
    double px = s.Pos.X, py = s.Pos.Y;
    double vx = s.Vel.X, vy = s.Vel.Y, mass = s.Mass;
    double r1_3, r2_3;
    gravity(px, py, &r1_3, &r2_3);
    double v = sqrt(vx*vx + vy*vy);
    double tmv = -th->Thrust / (mass * v);  /* negative: retrograde */
    double ax = ((2*W*vy + W*W*px) - mu1*(px-x1)/r1_3) - mu2*(px-x2)/r2_3 + tmv*vx;
    double ay = (-2*W*vx + W*W*py) - (mu1/r1_3+mu2/r2_3)*py + tmv*vy;
    *dPos = s.Vel;
    *dVel = (Vec){ ax, ay, 0 };
    *dm   = MassRate(*th);
}

/* --- Jacobi constant and events --- */

static double JacobiConstantEM(State s) {
    double v2 = v3Norm2(s.Vel);
    double d1 = hypot(s.Pos.X - x1, s.Pos.Y);
    double d2 = hypot(s.Pos.X - x2, s.Pos.Y);
    return 0.5*v2 - 0.5*W*W*(s.Pos.X*s.Pos.X + s.Pos.Y*s.Pos.Y) - mu1/d1 - mu2/d2;
}

static double EventJacobiEM(double t, State s, double phiS0, void *ctx) {
    double threshold = *(double *)ctx;
    return JacobiConstantEM(s) - threshold;
}

static double EventL1EM(double t, State s, double phiS0, void *ctx) {
    return hypot(s.Pos.X - x1, s.Pos.Y) - L1dist;
}

/* --- GTOInitialState --- */

static State GTOInitialState(GTOConfig *cfg) {
    double rApogee  = rEarth + cfg->HApogee;
    double rPerigee = rEarth + cfg->HPerigee;
    double e  = (rApogee - rPerigee) / (rApogee + rPerigee);
    double v0 = sqrt(mu1*(1-e)/rApogee) - W*rApogee;

    double sinPhi = sin(cfg->Phi), cosPhi = cos(cfg->Phi);
    double sinGam = sin(cfg->Gamma), cosGam = cos(cfg->Gamma);

    State s;
    s.Pos  = (Vec){ rApogee*cosPhi + x1, rApogee*sinPhi, 0 };
    s.Vel  = (Vec){ v0*(sinGam*cosPhi - cosGam*sinPhi),
                    v0*(sinGam*sinPhi + cosGam*cosPhi), 0 };
    s.Mass = cfg->M0;
    return s;
}

/* --- SelectInitialStep (matches Go's Hairer §II.4 algorithm) --- */

static double SelectInitialStep(Integrator *ig) {
    const double errOrder = 4.0;
    const double nComp    = 5.0;

    State s = ig->State;
    Vec dPos0, dVel0; double dm0;
    ig->Rates(ig->T, s, ig->PhiS0, &dPos0, &dVel0, &dm0, ig->RatesCtx);

    double atol = ig->ATol, rtol = ig->RTol;
    double sc[7] = {
        atol + fabs(s.Pos.X)*rtol, atol + fabs(s.Pos.Y)*rtol, atol + fabs(s.Pos.Z)*rtol,
        atol + fabs(s.Vel.X)*rtol, atol + fabs(s.Vel.Y)*rtol, atol + fabs(s.Vel.Z)*rtol,
        atol + fabs(s.Mass)*rtol,
    };

    #define rmsNorm7(px,py,pz,vx,vy,vz,m) \
        sqrt(((px)*(px)+(py)*(py)+(pz)*(pz)+(vx)*(vx)+(vy)*(vy)+(vz)*(vz)+(m)*(m))/nComp)

    double d0 = rmsNorm7(s.Pos.X/sc[0], s.Pos.Y/sc[1], s.Pos.Z/sc[2],
                         s.Vel.X/sc[3], s.Vel.Y/sc[4], s.Vel.Z/sc[5], s.Mass/sc[6]);
    double d1 = rmsNorm7(dPos0.X/sc[0], dPos0.Y/sc[1], dPos0.Z/sc[2],
                         dVel0.X/sc[3], dVel0.Y/sc[4], dVel0.Z/sc[5], dm0/sc[6]);

    double h0;
    if (d0 < 1e-5 || d1 < 1e-5)
        h0 = 1e-6;
    else
        h0 = 0.01 * d0 / d1;

    State s1;
    s1.Pos  = v3Add(s.Pos, v3Scale(h0, dPos0));
    s1.Vel  = v3Add(s.Vel, v3Scale(h0, dVel0));
    s1.Mass = s.Mass + h0*dm0;

    Vec dPos1, dVel1; double dm1;
    ig->Rates(ig->T + h0, s1, ig->PhiS0, &dPos1, &dVel1, &dm1, ig->RatesCtx);

    Vec ddPos = v3Sub(dPos1, dPos0), ddVel = v3Sub(dVel1, dVel0);
    double ddm = dm1 - dm0;
    double d2 = rmsNorm7(ddPos.X/sc[0], ddPos.Y/sc[1], ddPos.Z/sc[2],
                         ddVel.X/sc[3], ddVel.Y/sc[4], ddVel.Z/sc[5], ddm/sc[6]) / h0;

    #undef rmsNorm7

    double h1;
    double maxD = fmax(d1, d2);
    if (maxD <= 1e-5)
        h1 = fmax(1e-6, h0*1e-3);
    else
        h1 = pow(0.01/maxD, 1.0/(errOrder+1));

    return fmin(100*h0, h1);
}

/* --- Dormand-Prince RK45 coefficients --- */

static const double dpC[7] = { 0, 1.0/5, 3.0/10, 4.0/5, 8.0/9, 1, 1 };

static const double dpA[7][6] = {
    { 0 },
    { 1.0/5 },
    { 3.0/40, 9.0/40 },
    { 44.0/45, -56.0/15, 32.0/9 },
    { 19372.0/6561, -25360.0/2187, 64448.0/6561, -212.0/729 },
    { 9017.0/3168, -355.0/33, 46732.0/5247, 49.0/176, -5103.0/18656 },
    { 35.0/384, 0, 500.0/1113, 125.0/192, -2187.0/6784, 11.0/84 },
};

static const double dpB[7] = {
    35.0/384, 0, 500.0/1113, 125.0/192, -2187.0/6784, 11.0/84, 0
};

static const double dpE[7] = {
    71.0/57600, 0, -71.0/16695, 71.0/1920, -17253.0/339200, 22.0/525, -1.0/40,
};

/* --- Step: single RK45 adaptive step --- */

static double Step(Integrator *ig, double h) {
    const double safety = 0.9;
    const double minFac = 0.2;
    const double maxFac = 10.0;
    const double order  = 5.0;

    double    t     = ig->T;
    State     s     = ig->State;
    double    phiS0 = ig->PhiS0;
    RatesFunc rates = ig->Rates;
    void     *ctx   = ig->RatesCtx;

    int stepRejected = 0;
    for (;;) {
        double tNext = t + h;
        double hEff  = tNext - t;

        Vec    kPos[7], kVel[7];
        double kM[7];

        rates(t, s, phiS0, &kPos[0], &kVel[0], &kM[0], ctx);

        for (int i = 1; i < 7; i++) {
            Vec    sumPos  = {0, 0, 0};
            Vec    sumVel  = {0, 0, 0};
            double sumMass = 0;
            for (int j = 0; j < i; j++) {
                sumPos.X += dpA[i][j] * kPos[j].X;
                sumPos.Y += dpA[i][j] * kPos[j].Y;
                sumVel.X += dpA[i][j] * kVel[j].X;
                sumVel.Y += dpA[i][j] * kVel[j].Y;
                sumMass  += dpA[i][j] * kM[j];
            }
            double ti = t + dpC[i]*hEff;
            State si;
            si.Pos  = v3Add(s.Pos, v3Scale(hEff, sumPos));
            si.Vel  = v3Add(s.Vel, v3Scale(hEff, sumVel));
            si.Mass = s.Mass + hEff*sumMass;
            rates(ti, si, phiS0, &kPos[i], &kVel[i], &kM[i], ctx);
        }

        Vec    sumBPos = {0,0,0}, sumBVel = {0,0,0};
        Vec    sumEPos = {0,0,0}, sumEVel = {0,0,0};
        double sumBM = 0, sumEM = 0;
        for (int i = 0; i < 7; i++) {
            sumBPos.X += dpB[i] * kPos[i].X;
            sumBPos.Y += dpB[i] * kPos[i].Y;
            sumBVel.X += dpB[i] * kVel[i].X;
            sumBVel.Y += dpB[i] * kVel[i].Y;
            sumBM     += dpB[i] * kM[i];
            sumEPos.X += dpE[i] * kPos[i].X;
            sumEPos.Y += dpE[i] * kPos[i].Y;
            sumEVel.X += dpE[i] * kVel[i].X;
            sumEVel.Y += dpE[i] * kVel[i].Y;
            sumEM     += dpE[i] * kM[i];
        }

        Vec    newPos = v3Add(s.Pos, v3Scale(hEff, sumBPos));
        Vec    newVel = v3Add(s.Vel, v3Scale(hEff, sumBVel));
        double newM   = s.Mass + hEff*sumBM;
        Vec    errPos = v3Scale(hEff, sumEPos);
        Vec    errVel = v3Scale(hEff, sumEVel);
        double errM   = hEff * sumEM;

        /* Per-component error scaling, n=5 state components (x,y,vx,vy,m). */
        double scPX    = ig->ATol + ig->RTol*fmax(fabs(s.Pos.X), fabs(newPos.X));
        double scPY    = ig->ATol + ig->RTol*fmax(fabs(s.Pos.Y), fabs(newPos.Y));
        double scVX    = ig->ATol + ig->RTol*fmax(fabs(s.Vel.X), fabs(newVel.X));
        double scVY    = ig->ATol + ig->RTol*fmax(fabs(s.Vel.Y), fabs(newVel.Y));
        double scaleM  = ig->ATol + ig->RTol*fmax(fabs(s.Mass),  fabs(newM));

        double sumSq = (errPos.X/scPX)*(errPos.X/scPX) +
                       (errPos.Y/scPY)*(errPos.Y/scPY) +
                       (errVel.X/scVX)*(errVel.X/scVX) +
                       (errVel.Y/scVY)*(errVel.Y/scVY) +
                       (errM/scaleM)*(errM/scaleM);
        double errNorm = sqrt(sumSq) / sqrt(5.0);

        if (errNorm <= 1.0) {
            ig->T     = tNext;
            ig->State = (State){ newPos, newVel, newM };
            ig->LastErrNorm = errNorm;
            ig->StepCount++;

            if (errNorm == 0.0)
                return fmin(hEff*maxFac, ig->MaxStep);

            double factorRaw     = safety * pow(errNorm, -1.0/order);
            double factorClamped = fmax(minFac, fmin(maxFac, factorRaw));
            if (stepRejected) factorClamped = fmin(1.0, factorClamped);
            return fmax(ig->MinStep, fmin(ig->MaxStep, hEff*factorClamped));
        }

        /* Step rejected: reduce h. */
        double factor = safety * pow(errNorm, -1.0/order);
        factor = fmax(minFac, factor);
        h = fmax(ig->MinStep, hEff*factor);
        stepRejected = 1;

        if (h <= ig->MinStep) {
            ig->T     = tNext;
            ig->State = (State){ newPos, newVel, newM };
            return ig->MinStep;
        }
    }
}

/* --- bisectEvent --- */

static double bisectEvent(double t0, double t1, State s0, State s1,
                          double phiS0, EventFunc ev, void *evCtx) {
    const int    maxIter = 50;
    const double tol     = 1e-10;

    for (int i = 0; i < maxIter; i++) {
        double tm = 0.5 * (t0 + t1);
        if (t1 - t0 < tol)
            return tm;

        double alpha = (tm - t0) / (t1 - t0);
        State sm;
        sm.Pos  = v3Add(v3Scale(1-alpha, s0.Pos),  v3Scale(alpha, s1.Pos));
        sm.Vel  = v3Add(v3Scale(1-alpha, s0.Vel),  v3Scale(alpha, s1.Vel));
        sm.Mass = (1-alpha)*s0.Mass + alpha*s1.Mass;

        double vm = ev(tm, sm, phiS0, evCtx);
        double v0 = ev(t0, s0, phiS0, evCtx);

        if (v0*vm < 0) {
            t1 = tm; s1 = sm;
        } else {
            t0 = tm; s0 = sm;
        }
    }
    return 0.5 * (t0 + t1);
}

/* --- IntegrateUntil --- */

static int IntegrateUntil(Integrator *ig, double tf,
                          Event *events, int nevents, double *outTime) {
    double h = ig->MaxStep;

    double prevEv[8]; /* sufficient for this benchmark */
    for (int i = 0; i < nevents; i++)
        prevEv[i] = events[i].fn(ig->T, ig->State, ig->PhiS0, events[i].ctx);

    while (ig->T < tf) {
        if (ig->T + h > tf)
            h = tf - ig->T;

        double tPrev = ig->T;
        State  sPrev = ig->State;
        h = Step(ig, h);

        for (int i = 0; i < nevents; i++) {
            double curr = events[i].fn(ig->T, ig->State, ig->PhiS0, events[i].ctx);
            if (prevEv[i] * curr < 0) {
                double eventTime = bisectEvent(tPrev, ig->T, sPrev, ig->State,
                                              ig->PhiS0, events[i].fn, events[i].ctx);
                double alpha = (eventTime - tPrev) / (ig->T - tPrev);
                ig->State.Pos  = v3Add(v3Scale(1-alpha, sPrev.Pos),  v3Scale(alpha, ig->State.Pos));
                ig->State.Vel  = v3Add(v3Scale(1-alpha, sPrev.Vel),  v3Scale(alpha, ig->State.Vel));
                ig->State.Mass = (1-alpha)*sPrev.Mass + alpha*ig->State.Mass;
                ig->T  = eventTime;
                *outTime = eventTime;
                return i;
            }
            prevEv[i] = curr;
        }
    }
    *outTime = ig->T;
    return -1;
}

/* --- printState --- */

static void printState(int phase, double t, State s) {
    printf("phase=%d t=%6.2fd earthdist=%.1fkm pos=(%.3f,%.3f)\n",
           phase, t/days, v3Norm(s.Pos), s.Pos.X, s.Pos.Y);
}

/* --- run --- */

static int run(double lowerBoundStep, int verify) {
    GTOConfig cfg = {
        .HApogee   = 37000,
        .HPerigee  = 1200,
        .Phi       = 0.7505211952744961 * M_PI / 180.0,
        .Gamma     = 0,
        .M0        = 12,
        .Thruster  = { .Thrust = 4 * 0.000000450, .Isp = 1650 },
        .JacobiThr = -1.63907788,
        .Tol       = 1e-12,
    };

    Integrator ig = {
        .T        = 0,
        .State    = GTOInitialState(&cfg),
        .PhiS0    = 0,
        .Rates    = RatesThrustEM,
        .RatesCtx = &cfg.Thruster,
        .MinStep  = 0,
        .MaxStep  = fmin(450.0, lowerBoundStep),
        .ATol     = cfg.Tol,
        .RTol     = 1e-9,
    };

    if (verify)
        printState(0, 0, ig.State);

    /* Phase 1: prograde thrust until Jacobi threshold. */
    double jThr1 = cfg.JacobiThr;
    Event ev1 = { EventJacobiEM, &jThr1 };
    double t1;
    int evIdx = IntegrateUntil(&ig, days*360, &ev1, 1, &t1);
    if (evIdx < 0) {
        fprintf(stderr, "phase 1: jacobi threshold not crossed\n");
        return 1;
    }
    if (verify)
        printState(1, t1, ig.State);

    /* Phase 2: coast until distance from earth = L1. */
    ig.Rates    = RatesCoastEM;
    ig.RatesCtx = NULL;
    ig.MaxStep  = fmin(200.0, lowerBoundStep);
    Event ev2 = { EventL1EM, NULL };
    double t2;
    evIdx = IntegrateUntil(&ig, t1 + 260*days, &ev2, 1, &t2);
    if (evIdx < 0) {
        fprintf(stderr, "during phase 2: L1 distance not reached\n");
        return 1;
    }
    if (verify)
        printState(2, t2, ig.State);

    /* Phase 3: retrograde braking until Jacobi constant reaches C1. */
    ig.Rates    = RatesBrakeEM;
    ig.RatesCtx = &cfg.Thruster;
    ig.MaxStep  = fmin(1.0*days, lowerBoundStep);
    double c1val = C1;
    Event ev3 = { EventJacobiEM, &c1val };
    double t3;
    evIdx = IntegrateUntil(&ig, t2 + 25*days, &ev3, 1, &t3);
    if (evIdx < 0) {
        fprintf(stderr, "during phase 3: C1 distance not reached while braking\n");
        return 1;
    }
    if (verify)
        printState(3, t3, ig.State);

    /* Phase 4: 20-day coast, no events. */
    ig.Rates    = RatesCoastEM;
    ig.RatesCtx = NULL;
    ig.MaxStep  = fmin(100.0, lowerBoundStep);
    double t4;
    IntegrateUntil(&ig, t3 + 20*days, NULL, 0, &t4);
    if (verify)
        printState(4, t4, ig.State);

    return 0;
}

int main(int argc, char **argv) {
    if (argc < 2) {
        fprintf(stderr, "missing arg\n");
        return 1;
    }
    double lowerBoundStep = atof(argv[1]);
    int verify = argc == 3 && strcmp(argv[2], "v") == 0;

    /* Initialise runtime-derived constants (require sqrt or float arithmetic). */
    nS = sqrt((bigG * mS) / (rB2S * rB2S * rB2S));
    W  = sqrt((mu1 + mu2)  / (r12  * r12  * r12));
    x1 = -(m2 / (m1 + m2)) * r12;
    x2 =  (m1 / (m1 + m2)) * r12;

    return run(lowerBoundStep, verify);
}

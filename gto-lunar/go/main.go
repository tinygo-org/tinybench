// Original Author: Juan Pablo Sellanes Goncalves
// https://eventos.iua.edu.ar/event/1/contributions/51/

package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
)

// Physical constants, mostly relating to Earth-moon-sun system.
const (
	g0     = 9.807e-3    // [km/s²]
	days   = 24 * 3600   // [s/day]
	bigG   = 6.6742e-20  // [km³/kg/s²]
	r12    = 384400.0    // [km] Earth-Moon distance
	m1     = 5.974e24    // [kg] Earth mass
	m2     = 7.348e22    // [kg] Moon mass
	mu1    = 398600.0    // [km³/s²] Earth gravitational parameter
	mu2    = 4903.02     // [km³/s²] Moon gravitational parameter
	mS     = 1.989e30    // [kg] Sun mass
	rB2S   = 149597870.7 // [km] Barycenter to Sun distance
	rMoon  = 1737.0      // [km] Moon radius
	rEarth = 6378.0      // [km] Earth radius
	L1dist = 321710.0    // [km] L1 distance from Earth center
	C1     = -1.676      // Jacobi constant for braking phase

	mu  = mu1 + mu2      // [km³/s²] Combined gravitational parameter
	pi1 = m1 / (m1 + m2) // Earth mass fraction
	pi2 = m2 / (m1 + m2) // Moon mass fraction

	x1  = -pi2 * r12 // [km] Earth x-position in rotating frame
	x2  = pi1 * r12  // [km] Moon x-position in rotating frame
	muS = bigG * mS  // [km³/s²] Sun gravitational parameter
)

// Physical constants.
var (
	nS = math.Sqrt(muS / (rB2S * rB2S * rB2S)) // [rad/s] Sun mean motion
	W  = math.Sqrt(mu / (r12 * r12 * r12))     // [rad/s] Angular velocity of rotating frame
)

func main() {
	const nargs = 1
	if len(os.Args) < nargs+1 {
		fmt.Println("missing arg")
		os.Exit(1)
	}
	maxStep, err := strconv.ParseFloat(os.Args[1], 64)
	if err != nil {
		fmt.Println("bad arg:", err)
		os.Exit(1)
	}
	verify := len(os.Args) == 3 && os.Args[2] == "v"
	err = run(maxStep, verify)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(maxStep float64, verify bool) error {
	cfg := GTOConfig{
		HApogee:   37000,
		HPerigee:  1200,
		Phi:       0.7505211952744961 * math.Pi / 180,
		Gamma:     0,
		M0:        12,
		Thruster:  Thruster{Thrust: 4 * 0.000000450, Isp: 1650},
		JacobiThr: -1.63907788,
		Tol:       1e-12,
	}
	ig := Integrator{
		T:       0,
		State:   cfg.GTOInitialState(),
		PhiS0:   0,
		Rates:   RatesThrustEM(cfg.Thruster),
		MinStep: 0,
		MaxStep: min(450, maxStep),
		ATol:    cfg.Tol,
		RTol:    1e-9,
	}
	if verify {
		printState(0, 0, ig.State)
	}

	// Phase 1: prograde thrust until Jacobi threshold.
	evIdx, t1 := ig.IntegrateUntil(days*360, EventJacobiEM(cfg.JacobiThr))
	if evIdx < 0 {
		return errors.New("phase 1: jacobi threshold not crossed")
	}
	if verify {
		printState(1, t1, ig.State)
	}

	// Phase 2: coast until distance from earth = L1.
	ig.Rates = RatesCoastEM
	ig.MaxStep = min(200, maxStep)
	evIdx, t2 := ig.IntegrateUntil(t1+260*days, EventL1EM)
	if evIdx < 0 {
		return errors.New("during phase 2: L1 distance not reached")
	}
	if verify {
		printState(2, t2, ig.State)
	}

	// Phase 3: retrograde braking until Jacobi constant reaches C1.
	ig.Rates = RatesBrakeEM(cfg.Thruster)
	ig.MaxStep = min(1*days, maxStep)
	evIdx, t3 := ig.IntegrateUntil(t2+25*days, EventJacobiEM(C1))
	if evIdx < 0 {
		return errors.New("during phase 3: C1 distance not reached while braking")
	}
	if verify {
		printState(3, t3, ig.State)
	}

	// Phase 4: 20-day coast, no events.
	ig.Rates = RatesCoastEM
	ig.MaxStep = min(100, maxStep)
	_, t4 := ig.IntegrateUntil(t3 + 20*days)
	if verify {
		printState(4, t4, ig.State)
	}
	return nil
}

func printState(phase int, t float64, s State) {
	fmt.Printf("phase=%d t=%6.2fd earthdist=%.1fkm pos=(%.3f,%.3f)\n", phase, t/days, v3Norm(s.Pos), s.Pos.X, s.Pos.Y)
}

// GTOConfig holds initial conditions and simulation parameters for the
// GTO-to-lunar-capture trajectory (Earth+Moon CR3BP, no Sun).
type GTOConfig struct {
	HApogee   float64  // [km] apogee altitude above Earth surface
	HPerigee  float64  // [km] perigee altitude above Earth surface
	Phi       float64  // [rad] initial angle of apogee from rotating-frame x-axis
	Gamma     float64  // [rad] flight path angle (0 = tangential departure)
	M0        float64  // [kg] initial spacecraft mass
	Thruster  Thruster // reused from rewrite.go
	JacobiThr float64  // Jacobi constant threshold to end Phase 1
	Tol       float64  // ATol; RTol fixed at 1e-9
}

// GTOInitialState computes spacecraft state at GTO apogee in the rotating frame.
// Uses vis-viva at apogee: v = sqrt(mu1*(1-e)/rApogee) minus frame rotation W*rApogee.
func (cfg *GTOConfig) GTOInitialState() State {
	rApogee := rEarth + cfg.HApogee
	rPerigee := rEarth + cfg.HPerigee
	e := (rApogee - rPerigee) / (rApogee + rPerigee)
	v0 := math.Sqrt(mu1*(1-e)/rApogee) - W*rApogee

	sinPhi, cosPhi := math.Sincos(cfg.Phi)
	sinGam, cosGam := math.Sincos(cfg.Gamma)

	return State{
		Pos:  Vec{X: rApogee*cosPhi + x1, Y: rApogee * sinPhi},
		Vel:  Vec{X: v0 * (sinGam*cosPhi - cosGam*sinPhi), Y: v0 * (sinGam*sinPhi + cosGam*cosPhi)},
		Mass: cfg.M0,
	}
}

// SelectInitialStep computes a safe first step size following Hairer, Norsett
// & Wanner §II.4 — the same algorithm used by scipy's select_initial_step.
// Uses per-component scaling (sc_i = ATol + |y_i|·RTol) unlike Step() which
// uses vector-norm scaling; this matches scipy's step-size controller exactly.
func (ig *Integrator) SelectInitialStep() float64 {
	const (
		errOrder = 4.0 // RK45 error-estimator order (scipy: error_estimator_order)
		nComp    = 5.0 // active state components for this 2D problem: x,y,vx,vy,m
		// (z and vz are always 0; including them would dilute the rms by sqrt(5/7)
		// and shift the initial h ~3.4% away from scipy's value)
	)
	s := ig.State
	dPos0, dVel0, dm0 := ig.Rates(ig.T, s, ig.PhiS0)

	// Per-component scale factors.
	atol, rtol := ig.ATol, ig.RTol
	sc := [7]float64{
		atol + math.Abs(s.Pos.X)*rtol, atol + math.Abs(s.Pos.Y)*rtol, atol + math.Abs(s.Pos.Z)*rtol,
		atol + math.Abs(s.Vel.X)*rtol, atol + math.Abs(s.Vel.Y)*rtol, atol + math.Abs(s.Vel.Z)*rtol,
		atol + math.Abs(s.Mass)*rtol,
	}
	rmsNorm := func(px, py, pz, vx, vy, vz, m float64) float64 {
		return math.Sqrt((px*px + py*py + pz*pz + vx*vx + vy*vy + vz*vz + m*m) / nComp)
	}

	d0 := rmsNorm(s.Pos.X/sc[0], s.Pos.Y/sc[1], s.Pos.Z/sc[2],
		s.Vel.X/sc[3], s.Vel.Y/sc[4], s.Vel.Z/sc[5], s.Mass/sc[6])
	d1 := rmsNorm(dPos0.X/sc[0], dPos0.Y/sc[1], dPos0.Z/sc[2],
		dVel0.X/sc[3], dVel0.Y/sc[4], dVel0.Z/sc[5], dm0/sc[6])

	var h0 float64
	if d0 < 1e-5 || d1 < 1e-5 {
		h0 = 1e-6
	} else {
		h0 = 0.01 * d0 / d1
	}

	// One explicit Euler probe step to estimate second derivative.
	s1 := State{
		Pos:  v3Add(s.Pos, v3Scale(h0, dPos0)),
		Vel:  v3Add(s.Vel, v3Scale(h0, dVel0)),
		Mass: s.Mass + h0*dm0,
	}
	dPos1, dVel1, dm1 := ig.Rates(ig.T+h0, s1, ig.PhiS0)
	ddPos, ddVel, ddm := v3Sub(dPos1, dPos0), v3Sub(dVel1, dVel0), dm1-dm0
	d2 := rmsNorm(ddPos.X/sc[0], ddPos.Y/sc[1], ddPos.Z/sc[2],
		ddVel.X/sc[3], ddVel.Y/sc[4], ddVel.Z/sc[5], ddm/sc[6]) / h0

	var h1 float64
	if maxD := math.Max(d1, d2); maxD <= 1e-5 {
		h1 = math.Max(1e-6, h0*1e-3)
	} else {
		h1 = math.Pow(0.01/maxD, 1/(errOrder+1))
	}
	return math.Min(100*h0, h1)
}

// RatesThrustEM returns a RatesFunc for prograde thrust in Earth+Moon CR3BP.
//
//	ax = 2*W*vy + W²*x - mu1*(x-x1)/r1³ - mu2*(x-x2)/r2³ + (T/m)*(vx/v)
//	ay = -2*W*vx + W²*y - (mu1/r1³ + mu2/r2³)*y + (T/m)*(vy/v)
func RatesThrustEM(th Thruster) RatesFunc {
	return func(t float64, s State, _ float64) (dPos, dVel Vec, dm float64) {
		x, y := s.Pos.X, s.Pos.Y
		vx, vy, m := s.Vel.X, s.Vel.Y, s.Mass
		r1_3, r2_3 := gravity(x, y)
		v := math.Sqrt(vx*vx + vy*vy)
		tmv := th.Thrust / (m * v)
		ax := ((2*W*vy + W*W*x) - mu1*(x-x1)/r1_3) - mu2*(x-x2)/r2_3 + tmv*vx
		ay := (-2*W*vx + W*W*y) - (mu1/r1_3+mu2/r2_3)*y + tmv*vy
		dPos = s.Vel
		dVel = Vec{X: ax, Y: ay}
		dm = th.MassRate()
		return
	}
}

// RatesCoastEM is a RatesFunc for unpowered coasting in Earth+Moon CR3BP.
// Mirrors Python rates0(t,f) sequential left-to-right evaluation exactly.
func RatesCoastEM(t float64, s State, _ float64) (dPos, dVel Vec, dm float64) {
	x, y := s.Pos.X, s.Pos.Y
	vx, vy := s.Vel.X, s.Vel.Y
	r1_3, r2_3 := gravity(x, y)
	ax := ((2*W*vy + W*W*x) - mu1*(x-x1)/r1_3) - mu2*(x-x2)/r2_3
	ay := (-2*W*vx + W*W*y) - (mu1/r1_3+mu2/r2_3)*y
	dPos = s.Vel
	dVel = Vec{X: ax, Y: ay}
	return
}

// RatesBrakeEM returns a RatesFunc for retrograde braking in Earth+Moon CR3BP.
// Mirrors Python rates_1(t,f) sequential left-to-right evaluation exactly.
func RatesBrakeEM(th Thruster) RatesFunc {
	return func(t float64, s State, _ float64) (dPos, dVel Vec, dm float64) {
		x, y := s.Pos.X, s.Pos.Y
		vx, vy, m := s.Vel.X, s.Vel.Y, s.Mass
		r1_3, r2_3 := gravity(x, y)
		v := math.Sqrt(vx*vx + vy*vy)
		tmv := -th.Thrust / (m * v) // negative: retrograde
		ax := ((2*W*vy + W*W*x) - mu1*(x-x1)/r1_3) - mu2*(x-x2)/r2_3 + tmv*vx
		ay := (-2*W*vx + W*W*y) - (mu1/r1_3+mu2/r2_3)*y + tmv*vy
		dPos = s.Vel
		dVel = Vec{X: ax, Y: ay}
		dm = th.MassRate()
		return
	}
}

// EventJacobiEM returns an EventFunc that triggers when the EM-only Jacobi
// constant crosses threshold. Matches Python jacobiC and jacobiC1 behavior.
func EventJacobiEM(threshold float64) EventFunc {
	return func(t float64, s State, _ float64) float64 {
		return JacobiConstantEM(s) - threshold
	}
}

// EventL1EM triggers when the spacecraft's distance from Earth center equals
// L1dist (321710 km). Matches Python lagrian1(t, y).
func EventL1EM(t float64, s State, _ float64) float64 {
	return math.Hypot(s.Pos.X-x1, s.Pos.Y) - L1dist
}

// JacobiConstantEM computes the Jacobi constant for the Earth+Moon CR3BP (no Sun).
// J = 0.5v² - 0.5W²(x²+y²) - mu1/r1 - mu2/r2
func JacobiConstantEM(s State) float64 {
	v2 := v3Norm2(s.Vel)
	pos := s.Pos
	d1 := math.Hypot(pos.X-x1, pos.Y)
	d2 := math.Hypot(pos.X-x2, pos.Y)
	return 0.5*v2 - 0.5*W*W*(pos.X*pos.X+pos.Y*pos.Y) - mu1/d1 - mu2/d2
}

// gravity computes distances and per-body gravity coefficients used by
// all three rates functions. Matches Python's np.linalg.norm and r**3 style.
func gravity(x, y float64) (r1_3, r2_3 float64) {
	dx1 := x - x1
	dx2 := x - x2
	r1 := math.Sqrt(dx1*dx1 + y*y)
	r2 := math.Sqrt(dx2*dx2 + y*y)
	return r1 * r1 * r1, r2 * r2 * r2
}

type Vec struct {
	X, Y, Z float64
}

func v3Norm2(v Vec) float64        { return v.X*v.X + v.Y*v.Y + v.Z*v.Z }
func v3Norm(v Vec) float64         { return math.Hypot(v.X, math.Hypot(v.Y, v.Z)) }
func v3Sub(a, b Vec) Vec           { return Vec{X: a.X - b.X, Y: a.Y - b.Y, Z: a.Z - b.Z} }
func v3Add(a, b Vec) Vec           { return Vec{X: a.X + b.X, Y: a.Y + b.Y, Z: a.Z + b.Z} }
func v3Scale(f float64, v Vec) Vec { return Vec{X: f * v.X, Y: f * v.Y, Z: f * v.Z} }

// fma3 computes a*b+c using math.FMA per component.
func v3FMA(a float64, b, c Vec) Vec {
	return Vec{
		X: math.FMA(a, b.X, c.X),
		Y: math.FMA(a, b.Y, c.Y),
		Z: math.FMA(a, b.Z, c.Z),
	}
}

// State represents spacecraft state in the rotating Earth-Moon frame.
type State struct {
	Pos  Vec     // [km]
	Vel  Vec     // [km/s]
	Mass float64 // [kg]
}

// Thruster defines propulsion parameters.
type Thruster struct {
	Thrust float64 // [kN] Total thrust
	Isp    float64 // [s] Specific impulse
}

// MassRate returns mass flow rate [kg/s] (negative, mass decreases).
func (th Thruster) MassRate() float64 { return -th.Thrust / (g0 * th.Isp) }

// Trajectory holds initial conditions and computes transfer trajectory.
type Trajectory struct {
	D0    float64 // [km] Altitude above Earth
	Phi   float64 // [rad] Initial angle from Earth center
	Gamma float64 // [rad] Flight path angle
	PhiS0 float64 // [rad] Initial Sun angle in rotating frame
	M0    float64 // [kg] Initial mass

	Thruster  Thruster
	JacobiThr float64 // Jacobi threshold for phase 1 termination
	Tol       float64 // Integration tolerance
	MaxStep   float64 // [s] Maximum integration step
}

// InitialState computes the initial state from trajectory parameters.
func (tr *Trajectory) InitialState() State {
	r0 := rEarth + tr.D0
	v0 := math.Sqrt(mu1/r0) - W*r0 // Circular velocity minus frame rotation

	sinPhi, cosPhi := math.Sincos(tr.Phi)
	sinGam, cosGam := math.Sincos(tr.Gamma)

	return State{
		Pos: Vec{
			X: r0*cosPhi + x1,
			Y: r0 * sinPhi,
			Z: 0,
		},
		Vel: Vec{
			X: v0 * (sinGam*cosPhi - cosGam*sinPhi),
			Y: v0 * (sinGam*sinPhi + cosGam*cosPhi),
			Z: 0,
		},
		Mass: tr.M0,
	}
}

// RatesFunc computes state derivatives: dPos/dt, dVel/dt, dm/dt.
type RatesFunc func(t float64, s State, phiS0 float64) (dPos, dVel Vec, dm float64)

// EventFunc returns a value that crosses zero at an event.
type EventFunc func(t float64, s State, phiS0 float64) float64

// Integrator performs RK45 (Dormand-Prince) integration for spacecraft trajectories.
// It should match scipy integrator logic.
type Integrator struct {
	T     float64
	State State
	PhiS0 float64
	Rates RatesFunc

	// Step control
	MinStep, MaxStep float64
	ATol, RTol       float64

	// Diagnostics
	LastErrNorm float64 // error norm from most recent accepted step
	StepCount   int     // incremented on each accepted step
}

// Dormand-Prince coefficients (RK45)
var (
	// Nodes
	dpC = [7]float64{0, 1.0 / 5, 3.0 / 10, 4.0 / 5, 8.0 / 9, 1, 1}

	// Matrix A (lower triangular)
	dpA = [7][6]float64{
		{},
		{1.0 / 5},
		{3.0 / 40, 9.0 / 40},
		{44.0 / 45, -56.0 / 15, 32.0 / 9},
		{19372.0 / 6561, -25360.0 / 2187, 64448.0 / 6561, -212.0 / 729},
		{9017.0 / 3168, -355.0 / 33, 46732.0 / 5247, 49.0 / 176, -5103.0 / 18656},
		{35.0 / 384, 0, 500.0 / 1113, 125.0 / 192, -2187.0 / 6784, 11.0 / 84},
	}

	// 5th order weights
	dpB = [7]float64{35.0 / 384, 0, 500.0 / 1113, 125.0 / 192, -2187.0 / 6784, 11.0 / 84, 0}

	// Error estimation weights — direct form matching scipy's RK45.E coefficients.
	// Using direct rationals avoids catastrophic cancellation from the b-b* form.
	dpE = [7]float64{
		71.0 / 57600,
		0,
		-71.0 / 16695,
		71.0 / 1920,
		-17253.0 / 339200,
		22.0 / 525,
		-1.0 / 40,
	}
)

// Step performs a single RK45 step with adaptive step size.
// Returns the new suggested step size.
func (ig *Integrator) Step(h float64) float64 {
	const (
		safety = 0.9
		minFac = 0.2
		maxFac = 10.0
		order  = 5.0
	)

	t := ig.T
	s := ig.State
	phiS0 := ig.PhiS0
	rates := ig.Rates

	stepRejected := false
	for {
		// Mirror scipy's _step_impl inner loop: recompute h as (t+h)-t.
		// scipy does: t_new = t + h; h = t_new - t; h_abs = |h|
		// This round-trip can lose 1 ULP when t and h have similar magnitudes.
		// All stage computations and the factor scaling use this hEff, not the
		// original h. ig.T is set to tNext (= t + h_orig) to preserve the exact
		// accumulated time — scipy does self.t = t_new for the same reason.
		tNext := t + h
		hEff := tNext - t

		// Compute k values using hEff
		var kPos, kVel [7]Vec
		var kM [7]float64

		kPos[0], kVel[0], kM[0] = rates(t, s, phiS0)

		for i := 1; i < 7; i++ {
			// Match numpy: dy = dot(A[i,:i], K[:i]) * h  (standard += to match numpy C-loop)
			var sumPos, sumVel Vec
			var sumMass float64
			for j := 0; j < i; j++ {
				sumPos.X += dpA[i][j] * kPos[j].X
				sumPos.Y += dpA[i][j] * kPos[j].Y
				sumVel.X += dpA[i][j] * kVel[j].X
				sumVel.Y += dpA[i][j] * kVel[j].Y
				sumMass += dpA[i][j] * kM[j]
			}
			ti := t + dpC[i]*hEff
			si := State{
				Pos:  v3Add(s.Pos, v3Scale(hEff, sumPos)),
				Vel:  v3Add(s.Vel, v3Scale(hEff, sumVel)),
				Mass: s.Mass + hEff*sumMass,
			}
			kPos[i], kVel[i], kM[i] = rates(ti, si, phiS0)
		}

		// Compute 5th order solution and error estimate.
		// Match numpy: y_new = y + h*dot(B,K),  err = h*dot(E,K)
		// Use standard += (two rounds per term) to match numpy's non-FMA C-loop accumulation.
		var sumBPos, sumBVel, sumEPos, sumEVel Vec
		var sumBM, sumEM float64
		for i := 0; i < 7; i++ {
			sumBPos.X += dpB[i] * kPos[i].X
			sumBPos.Y += dpB[i] * kPos[i].Y
			sumBVel.X += dpB[i] * kVel[i].X
			sumBVel.Y += dpB[i] * kVel[i].Y
			sumBM += dpB[i] * kM[i]
			sumEPos.X += dpE[i] * kPos[i].X
			sumEPos.Y += dpE[i] * kPos[i].Y
			sumEVel.X += dpE[i] * kVel[i].X
			sumEVel.Y += dpE[i] * kVel[i].Y
			sumEM += dpE[i] * kM[i]
		}
		newPos := v3Add(s.Pos, v3Scale(hEff, sumBPos))
		newVel := v3Add(s.Vel, v3Scale(hEff, sumBVel))
		newM := s.Mass + hEff*sumBM
		errPos := v3Scale(hEff, sumEPos)
		errVel := v3Scale(hEff, sumEVel)
		errM := hEff * sumEM

		// Error norm — per-component scaling, n=5 state components (x,y,vx,vy,m).
		// All Python models use a 5-component 2D state; z and vz are always zero
		// so their error contributions are exactly zero and would only dilute the
		// denominator (7 vs 5), inflating step sizes by sqrt(5/7)≈15%.
		scPX := ig.ATol + ig.RTol*math.Max(math.Abs(s.Pos.X), math.Abs(newPos.X))
		scPY := ig.ATol + ig.RTol*math.Max(math.Abs(s.Pos.Y), math.Abs(newPos.Y))
		scVX := ig.ATol + ig.RTol*math.Max(math.Abs(s.Vel.X), math.Abs(newVel.X))
		scVY := ig.ATol + ig.RTol*math.Max(math.Abs(s.Vel.Y), math.Abs(newVel.Y))
		scaleM := ig.ATol + ig.RTol*math.Max(math.Abs(s.Mass), math.Abs(newM))

		// Match scipy: rms_norm(v) = np.linalg.norm(v) / sqrt(n)
		// = sqrt(sum(v_i^2)) / sqrt(5)  NOT sqrt(sum/5) — differ by 1 ULP.
		sumSq := (errPos.X/scPX)*(errPos.X/scPX) +
			(errPos.Y/scPY)*(errPos.Y/scPY) +
			(errVel.X/scVX)*(errVel.X/scVX) +
			(errVel.Y/scVY)*(errVel.Y/scVY) +
			(errM/scaleM)*(errM/scaleM)
		errNorm := math.Sqrt(sumSq) / math.Sqrt(5)

		// Step accepted?
		if errNorm <= 1 {
			ig.T = tNext
			ig.State = State{Pos: newPos, Vel: newVel, Mass: newM}
			ig.LastErrNorm = errNorm
			ig.StepCount++

			if errNorm == 0 {
				return math.Min(hEff*maxFac, ig.MaxStep)
			}
			factorRaw := safety * math.Pow(errNorm, -1.0/order)
			factorClamped := math.Max(minFac, math.Min(maxFac, factorRaw))
			// Mirror scipy: after rejection-then-acceptance, cap factor at 1.0
			// (scipy: if step_rejected: factor = min(1, factor))
			if stepRejected {
				factorClamped = math.Min(1.0, factorClamped)
			}
			return math.Max(ig.MinStep, math.Min(ig.MaxStep, hEff*factorClamped))
		}

		// Step rejected: reduce using hEff (mirrors scipy: h_abs *= factor where h_abs = |hEff|)
		factor := safety * math.Pow(errNorm, -1.0/order)
		factor = math.Max(minFac, factor)
		h = math.Max(ig.MinStep, hEff*factor)
		stepRejected = true

		if h <= ig.MinStep {
			ig.T = tNext
			ig.State = State{Pos: newPos, Vel: newVel, Mass: newM}
			return ig.MinStep
		}
	}
}

// IntegrateUntil integrates until tf or an event triggers.
// Returns the index of triggered event (-1 if none) and final time.
func (ig *Integrator) IntegrateUntil(tf float64, events ...EventFunc) (eventIdx int, eventTime float64) {
	h := ig.MaxStep
	eventIdx = -1

	// Evaluate events at start
	prevEvents := make([]float64, len(events))
	for i, ev := range events {
		prevEvents[i] = ev(ig.T, ig.State, ig.PhiS0)
	}

	for ig.T < tf {
		// Limit step to not overshoot
		if ig.T+h > tf {
			h = tf - ig.T
		}

		tPrev := ig.T
		sPrev := ig.State
		h = ig.Step(h)

		// Check events for sign change
		for i, ev := range events {
			curr := ev(ig.T, ig.State, ig.PhiS0)
			if prevEvents[i]*curr < 0 {
				// Sign change detected, find root with bisection
				eventTime = bisectEvent(tPrev, ig.T, sPrev, ig.State, ig.PhiS0, ev)
				eventIdx = i

				// Interpolate state at event time (linear approximation)
				alpha := (eventTime - tPrev) / (ig.T - tPrev)
				ig.State = State{
					Pos:  v3Add(v3Scale(1-alpha, sPrev.Pos), v3Scale(alpha, ig.State.Pos)),
					Vel:  v3Add(v3Scale(1-alpha, sPrev.Vel), v3Scale(alpha, ig.State.Vel)),
					Mass: (1-alpha)*sPrev.Mass + alpha*ig.State.Mass,
				}
				ig.T = eventTime
				return eventIdx, eventTime
			}
			prevEvents[i] = curr
		}
	}

	return -1, ig.T
}

// bisectEvent finds the time of event crossing using bisection.
func bisectEvent(t0, t1 float64, s0, s1 State, phiS0 float64, ev EventFunc) float64 {
	const maxIter = 50
	const tol = 1e-10

	for i := 0; i < maxIter; i++ {
		tm := 0.5 * (t0 + t1)
		if t1-t0 < tol {
			return tm
		}

		// Linear interpolation of state
		alpha := (tm - t0) / (t1 - t0)
		sm := State{
			Pos:  v3Add(v3Scale(1-alpha, s0.Pos), v3Scale(alpha, s1.Pos)),
			Vel:  v3Add(v3Scale(1-alpha, s0.Vel), v3Scale(alpha, s1.Vel)),
			Mass: (1-alpha)*s0.Mass + alpha*s1.Mass,
		}

		vm := ev(tm, sm, phiS0)
		v0 := ev(t0, s0, phiS0)

		if v0*vm < 0 {
			t1 = tm
			s1 = sm
		} else {
			t0 = tm
			s0 = sm
		}
	}

	return 0.5 * (t0 + t1)
}

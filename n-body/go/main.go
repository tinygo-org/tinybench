package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
)

const (
	pi          = 3.141592653589793
	solarMass   = 4 * pi * pi
	daysPerYear = 365.24
)

type Planet struct {
	x, y, z    float64
	vx, vy, vz float64
	mass       float64
}

func advance(nbodies int, bodies []Planet, dt float64) {
	for i := 0; i < nbodies; i++ {
		b := &bodies[i]
		for j := i + 1; j < nbodies; j++ {
			b2 := &bodies[j]
			dx := b.x - b2.x
			dy := b.y - b2.y
			dz := b.z - b2.z
			distanceSquared := dx*dx + dy*dy + dz*dz
			distance := math.Sqrt(distanceSquared)
			mag := dt / (distanceSquared * distance)
			b.vx -= dx * b2.mass * mag
			b.vy -= dy * b2.mass * mag
			b.vz -= dz * b2.mass * mag
			b2.vx += dx * b.mass * mag
			b2.vy += dy * b.mass * mag
			b2.vz += dz * b.mass * mag
		}
	}
	for i := 0; i < nbodies; i++ {
		b := &bodies[i]
		b.x += dt * b.vx
		b.y += dt * b.vy
		b.z += dt * b.vz
	}
}

func energy(nbodies int, bodies []Planet) float64 {
	e := 0.0
	for i := 0; i < nbodies; i++ {
		b := &bodies[i]
		e += 0.5 * b.mass * (b.vx*b.vx + b.vy*b.vy + b.vz*b.vz)
		for j := i + 1; j < nbodies; j++ {
			b2 := &bodies[j]
			dx := b.x - b2.x
			dy := b.y - b2.y
			dz := b.z - b2.z
			distance := math.Sqrt(dx*dx + dy*dy + dz*dz)
			e -= (b.mass * b2.mass) / distance
		}
	}
	return e
}

func offsetMomentum(nbodies int, bodies []Planet) {
	px, py, pz := 0.0, 0.0, 0.0
	for i := 0; i < nbodies; i++ {
		px += bodies[i].vx * bodies[i].mass
		py += bodies[i].vy * bodies[i].mass
		pz += bodies[i].vz * bodies[i].mass
	}
	bodies[0].vx = -px / solarMass
	bodies[0].vy = -py / solarMass
	bodies[0].vz = -pz / solarMass
}

const nbodies = 5

var bodies = [nbodies]Planet{
	{ // sun
		0, 0, 0, 0, 0, 0, solarMass,
	},
	{ // jupiter
		4.84143144246472090e+00,
		-1.16032004402742839e+00,
		-1.03622044471123109e-01,
		1.66007664274403694e-03 * daysPerYear,
		7.69901118419740425e-03 * daysPerYear,
		-6.90460016972063023e-05 * daysPerYear,
		9.54791938424326609e-04 * solarMass,
	},
	{ // saturn
		8.34336671824457987e+00,
		4.12479856412430479e+00,
		-4.03523417114321381e-01,
		-2.76742510726862411e-03 * daysPerYear,
		4.99852801234917238e-03 * daysPerYear,
		2.30417297573763929e-05 * daysPerYear,
		2.85885980666130812e-04 * solarMass,
	},
	{ // uranus
		1.28943695621391310e+01,
		-1.51111514016986312e+01,
		-2.23307578892655734e-01,
		2.96460137564761618e-03 * daysPerYear,
		2.37847173959480950e-03 * daysPerYear,
		-2.96589568540237556e-05 * daysPerYear,
		4.36624404335156298e-05 * solarMass,
	},
	{ // neptune
		1.53796971148509165e+01,
		-2.59193146099879641e+01,
		1.79258772950371181e-01,
		2.68067772490389322e-03 * daysPerYear,
		1.62824170038242295e-03 * daysPerYear,
		-9.51592254519715870e-05 * daysPerYear,
		5.15138902046611451e-05 * solarMass,
	},
}

func main() {
	n, _ := strconv.Atoi(os.Args[1])
	offsetMomentum(nbodies, bodies[:])
	fmt.Printf("%.9f\n", energy(nbodies, bodies[:]))
	for i := 1; i <= n; i++ {
		advance(nbodies, bodies[:], 0.01)
	}
	fmt.Printf("%.9f\n", energy(nbodies, bodies[:]))
}

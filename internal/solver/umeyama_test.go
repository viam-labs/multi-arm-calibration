package solver

import (
	"math"
	"math/rand"
	"testing"

	"github.com/golang/geo/r3"
	"go.viam.com/rdk/spatialmath"
	"go.viam.com/test"
)

func applyPose(p spatialmath.Pose, v r3.Vector) r3.Vector {
	tp := spatialmath.NewPoseFromPoint(v)
	out := spatialmath.Compose(p, tp)
	pt := out.Point()
	return r3.Vector{X: pt.X, Y: pt.Y, Z: pt.Z}
}

func randomPose(rng *rand.Rand) spatialmath.Pose {
	trans := r3.Vector{
		X: rng.Float64()*2000 - 1000,
		Y: rng.Float64()*2000 - 1000,
		Z: rng.Float64()*2000 - 1000,
	}
	ov := &spatialmath.OrientationVectorDegrees{
		OX:    rng.Float64()*2 - 1,
		OY:    rng.Float64()*2 - 1,
		OZ:    rng.Float64()*2 - 1,
		Theta: rng.Float64() * 360,
	}
	return spatialmath.NewPose(trans, ov)
}

func TestSolveRoundTrip(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	for range 20 {
		want := randomPose(rng)
		from := []r3.Vector{
			{X: 0, Y: 0, Z: 0},
			{X: 100, Y: 0, Z: 0},
			{X: 0, Y: 100, Z: 0},
			{X: 0, Y: 0, Z: 100},
			{X: 50, Y: 50, Z: 50},
		}
		to := make([]r3.Vector, len(from))
		for i, p := range from {
			to[i] = applyPose(want, p)
		}

		got, err := Solve(from, to)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, spatialmath.PoseAlmostCoincidentEps(got, want, 1e-6), test.ShouldBeTrue)
	}
}

func TestSolveIdentity(t *testing.T) {
	pts := []r3.Vector{
		{X: 1, Y: 0, Z: 0},
		{X: 0, Y: 1, Z: 0},
		{X: 0, Y: 0, Z: 1},
		{X: 1, Y: 1, Z: 1},
	}
	got, err := Solve(pts, pts)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, spatialmath.PoseAlmostCoincidentEps(got, spatialmath.NewZeroPose(), 1e-6), test.ShouldBeTrue)
}

func TestSolveRejectsTooFewPoints(t *testing.T) {
	from := []r3.Vector{{X: 0}, {X: 1}}
	to := []r3.Vector{{X: 0}, {X: 1}}
	_, err := Solve(from, to)
	test.That(t, err, test.ShouldEqual, ErrTooFewPoints)
}

func TestSolveRejectsLengthMismatch(t *testing.T) {
	from := []r3.Vector{{X: 0}, {X: 1}, {X: 2}}
	to := []r3.Vector{{X: 0}, {X: 1}}
	_, err := Solve(from, to)
	test.That(t, err, test.ShouldEqual, ErrLengthMismatch)
}

func TestSolveRejectsCollinearInput(t *testing.T) {
	from := []r3.Vector{
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 0},
		{X: 2, Y: 0, Z: 0},
		{X: 3, Y: 0, Z: 0},
	}
	to := []r3.Vector{
		{X: 0, Y: 10, Z: 0},
		{X: 1, Y: 10, Z: 0},
		{X: 2, Y: 10, Z: 0},
		{X: 3, Y: 10, Z: 0},
	}
	_, err := Solve(from, to)
	test.That(t, err, test.ShouldEqual, ErrDegenerate)
}

func TestSolveNoisyInputBoundedResidual(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	want := randomPose(rng)
	from := []r3.Vector{
		{X: 0, Y: 0, Z: 0},
		{X: 100, Y: 0, Z: 0},
		{X: 0, Y: 100, Z: 0},
		{X: 0, Y: 0, Z: 100},
		{X: 50, Y: 50, Z: 50},
	}
	const noiseMM = 0.1
	to := make([]r3.Vector, len(from))
	for i, p := range from {
		clean := applyPose(want, p)
		to[i] = r3.Vector{
			X: clean.X + (rng.Float64()*2-1)*noiseMM,
			Y: clean.Y + (rng.Float64()*2-1)*noiseMM,
			Z: clean.Z + (rng.Float64()*2-1)*noiseMM,
		}
	}

	got, err := Solve(from, to)
	test.That(t, err, test.ShouldBeNil)

	var maxResidual float64
	for i, p := range from {
		predicted := applyPose(got, p)
		d := math.Sqrt(math.Pow(predicted.X-to[i].X, 2) + math.Pow(predicted.Y-to[i].Y, 2) + math.Pow(predicted.Z-to[i].Z, 2))
		if d > maxResidual {
			maxResidual = d
		}
	}
	test.That(t, maxResidual, test.ShouldBeLessThan, noiseMM*3)
}

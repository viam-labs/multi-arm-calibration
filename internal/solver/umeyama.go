// Package solver implements the rigid-transform math for touch-based calibration.
package solver

import (
	"errors"
	"fmt"

	"github.com/golang/geo/r3"
	"go.viam.com/rdk/spatialmath"
	"gonum.org/v1/gonum/mat"
)

var (
	ErrTooFewPoints   = errors.New("need at least 3 correspondences")
	ErrLengthMismatch = errors.New("from and to must have the same length")
	ErrDegenerate     = errors.New("input points are degenerate (collinear or coincident)")
)

// Solve returns the rigid transform T such that T·from[i] ≈ to[i] for all i,
// via the Kabsch/Umeyama SVD algorithm (Umeyama 1991).
//
// TODO: consider upstreaming to spatialmath as RigidTransformFromCorrespondences —
// sanding hand-rolls the same math inline (with a reflection-correction bug) in
// cli/compare_mesh.go's ICP loop.
func Solve(from, to []r3.Vector) (spatialmath.Pose, error) {
	if len(from) != len(to) {
		return nil, ErrLengthMismatch
	}
	n := len(from)
	if n < 3 {
		return nil, ErrTooFewPoints
	}

	muFrom := centroid(from)
	muTo := centroid(to)

	h := mat.NewDense(3, 3, nil)
	for i := range n {
		f := from[i].Sub(muFrom)
		g := to[i].Sub(muTo)
		h.Set(0, 0, h.At(0, 0)+f.X*g.X)
		h.Set(0, 1, h.At(0, 1)+f.X*g.Y)
		h.Set(0, 2, h.At(0, 2)+f.X*g.Z)
		h.Set(1, 0, h.At(1, 0)+f.Y*g.X)
		h.Set(1, 1, h.At(1, 1)+f.Y*g.Y)
		h.Set(1, 2, h.At(1, 2)+f.Y*g.Z)
		h.Set(2, 0, h.At(2, 0)+f.Z*g.X)
		h.Set(2, 1, h.At(2, 1)+f.Z*g.Y)
		h.Set(2, 2, h.At(2, 2)+f.Z*g.Z)
	}

	var svd mat.SVD
	if ok := svd.Factorize(h, mat.SVDFull); !ok {
		return nil, fmt.Errorf("SVD failed")
	}
	if svd.Values(nil)[2] < 1e-9 {
		return nil, ErrDegenerate
	}

	var u, v mat.Dense
	svd.UTo(&u)
	svd.VTo(&v)

	var vut mat.Dense
	vut.Mul(&v, u.T())
	sign := 1.0
	if mat.Det(&vut) < 0 {
		sign = -1.0
	}

	var tmp, rMat mat.Dense
	tmp.Mul(&v, mat.NewDiagDense(3, []float64{1, 1, sign}))
	rMat.Mul(&tmp, u.T())

	rot, err := spatialmath.NewRotationMatrix([]float64{
		rMat.At(0, 0), rMat.At(0, 1), rMat.At(0, 2),
		rMat.At(1, 0), rMat.At(1, 1), rMat.At(1, 2),
		rMat.At(2, 0), rMat.At(2, 1), rMat.At(2, 2),
	})
	if err != nil {
		return nil, fmt.Errorf("rotation matrix: %w", err)
	}

	rotOnly := spatialmath.NewPose(r3.Vector{}, rot)
	rMuFrom := spatialmath.TransformPointByPose(rotOnly, muFrom)
	return spatialmath.NewPose(muTo.Sub(rMuFrom), rot), nil
}

// TODO: replace with spatialmath.Centroid once we bump RDK past its export.
func centroid(pts []r3.Vector) r3.Vector {
	var s r3.Vector
	for _, p := range pts {
		s = s.Add(p)
	}
	return s.Mul(1.0 / float64(len(pts)))
}

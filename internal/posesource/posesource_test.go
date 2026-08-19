package posesource

import (
	"context"
	"testing"

	"github.com/golang/geo/r3"
	"go.viam.com/rdk/spatialmath"
	"go.viam.com/test"
)

func poseAt(x, y, z float64) spatialmath.Pose {
	return spatialmath.NewPoseFromPoint(r3.Vector{X: x, Y: y, Z: z})
}

func TestFakeReturnsSequentialPoses(t *testing.T) {
	f := NewFake(map[string][]spatialmath.Pose{
		"arm_a": {poseAt(1, 0, 0), poseAt(2, 0, 0), poseAt(3, 0, 0)},
	})

	for i := 1; i <= 3; i++ {
		p, err := f.Capture(context.Background(), "arm_a")
		test.That(t, err, test.ShouldBeNil)
		test.That(t, p.Point().X, test.ShouldEqual, float64(i))
	}
}

func TestFakeRejectsUnknownArm(t *testing.T) {
	f := NewFake(map[string][]spatialmath.Pose{"arm_a": {poseAt(0, 0, 0)}})
	_, err := f.Capture(context.Background(), "arm_b")
	test.That(t, err, test.ShouldNotBeNil)
	test.That(t, err.Error(), test.ShouldContainSubstring, `unknown arm "arm_b"`)
}

func TestFakeRejectsExhaustedArm(t *testing.T) {
	f := NewFake(map[string][]spatialmath.Pose{"arm_a": {poseAt(1, 0, 0)}})
	_, err := f.Capture(context.Background(), "arm_a")
	test.That(t, err, test.ShouldBeNil)

	_, err = f.Capture(context.Background(), "arm_a")
	test.That(t, err, test.ShouldNotBeNil)
	test.That(t, err.Error(), test.ShouldContainSubstring, "exhausted")
}

func TestFakeInterleavesArmsIndependently(t *testing.T) {
	f := NewFake(map[string][]spatialmath.Pose{
		"arm_a": {poseAt(10, 0, 0), poseAt(20, 0, 0)},
		"arm_b": {poseAt(0, 10, 0), poseAt(0, 20, 0)},
	})

	pa1, _ := f.Capture(context.Background(), "arm_a")
	pb1, _ := f.Capture(context.Background(), "arm_b")
	pa2, _ := f.Capture(context.Background(), "arm_a")
	pb2, _ := f.Capture(context.Background(), "arm_b")

	test.That(t, pa1.Point().X, test.ShouldEqual, 10.0)
	test.That(t, pa2.Point().X, test.ShouldEqual, 20.0)
	test.That(t, pb1.Point().Y, test.ShouldEqual, 10.0)
	test.That(t, pb2.Point().Y, test.ShouldEqual, 20.0)
}

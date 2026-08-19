package touch

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/geo/r3"
	"go.viam.com/rdk/spatialmath"
	"go.viam.com/test"

	"github.com/viam-labs/multi-arm-calibration/internal/posdriver"
	"github.com/viam-labs/multi-arm-calibration/internal/posesource"
)

func poseAtPoint(x, y, z float64) spatialmath.Pose {
	return spatialmath.NewPoseFromPoint(r3.Vector{X: x, Y: y, Z: z})
}

func TestCalibrateRecoversKnownTransform(t *testing.T) {
	ctx := context.Background()
	cfg := validConfig()

	wantT := spatialmath.NewPose(
		r3.Vector{X: 500, Y: 200, Z: 0},
		&spatialmath.OrientationVectorDegrees{OZ: 1, Theta: 45},
	)

	armATCPs := []spatialmath.Pose{
		poseAtPoint(100, 0, 0),
		poseAtPoint(0, 100, 0),
		poseAtPoint(0, 0, 100),
		poseAtPoint(50, 50, 50),
	}
	// wantT is arm-2's pose relative to arm-1. Same physical point P has
	// coords P_a1 in arm-1's frame and P_a2 = wantT⁻¹ · P_a1 in arm-2's.
	inv := spatialmath.PoseInverse(wantT)
	armBTCPs := make([]spatialmath.Pose, len(armATCPs))
	for i, pose := range armATCPs {
		armBTCPs[i] = spatialmath.NewPoseFromPoint(spatialmath.TransformPointByPose(inv, pose.Point()))
	}

	p := posesource.NewFake(map[string][]spatialmath.Pose{
		"arm_a": armATCPs,
		"arm_b": armBTCPs,
	})
	d := posdriver.NewFake(
		"arm_a_home", "arm_a_p1", "arm_a_p2", "arm_a_p3", "arm_a_p4",
		"arm_b_home", "arm_b_p1", "arm_b_p2", "arm_b_p3", "arm_b_p4",
	)

	gotT, err := Calibrate(ctx, cfg, d, p)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, spatialmath.PoseAlmostCoincidentEps(gotT, wantT, 1e-6), test.ShouldBeTrue)
}

func TestCalibrateDrivesArmsInOrder(t *testing.T) {
	ctx := context.Background()
	cfg := validConfig()

	captures := []spatialmath.Pose{
		poseAtPoint(100, 0, 0), poseAtPoint(0, 100, 0),
		poseAtPoint(0, 0, 100), poseAtPoint(50, 50, 50),
	}
	p := posesource.NewFake(map[string][]spatialmath.Pose{
		"arm_a": captures,
		"arm_b": captures,
	})
	d := posdriver.NewFake(
		"arm_a_home", "arm_a_p1", "arm_a_p2", "arm_a_p3", "arm_a_p4",
		"arm_b_home", "arm_b_p1", "arm_b_p2", "arm_b_p3", "arm_b_p4",
	)

	_, err := Calibrate(ctx, cfg, d, p)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, d.Calls, test.ShouldResemble, []string{
		"arm_a_home", "arm_a_p1", "arm_a_p2", "arm_a_p3", "arm_a_p4", "arm_a_home",
		"arm_b_home", "arm_b_p1", "arm_b_p2", "arm_b_p3", "arm_b_p4", "arm_b_home",
	})
}

func TestCalibratePropagatesDriveError(t *testing.T) {
	ctx := context.Background()
	cfg := validConfig()

	p := posesource.NewFake(map[string][]spatialmath.Pose{
		"arm_a": {poseAtPoint(100, 0, 0)},
	})
	d := posdriver.NewFake("arm_a_home", "arm_a_p1", "arm_a_p2", "arm_a_p3", "arm_a_p4")
	boom := errors.New("boom")
	d.Errors["arm_a_p2"] = boom

	_, err := Calibrate(ctx, cfg, d, p)
	test.That(t, err, test.ShouldNotBeNil)
	test.That(t, errors.Is(err, boom), test.ShouldBeTrue)
}

func TestCalibratePropagatesCaptureError(t *testing.T) {
	ctx := context.Background()
	cfg := validConfig()

	p := posesource.NewFake(map[string][]spatialmath.Pose{
		"arm_a": {poseAtPoint(100, 0, 0)},
	})
	d := posdriver.NewFake(
		"arm_a_home", "arm_a_p1", "arm_a_p2", "arm_a_p3", "arm_a_p4",
		"arm_b_home", "arm_b_p1", "arm_b_p2", "arm_b_p3", "arm_b_p4",
	)

	_, err := Calibrate(ctx, cfg, d, p)
	test.That(t, err, test.ShouldNotBeNil)
	test.That(t, err.Error(), test.ShouldContainSubstring, "capture at saver[1]")
}

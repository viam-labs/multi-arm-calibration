package verifier

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/geo/r3"
	"go.viam.com/rdk/spatialmath"
	"go.viam.com/test"
)

func poseAt(x, y, z float64) spatialmath.Pose {
	return spatialmath.NewPoseFromPoint(r3.Vector{X: x, Y: y, Z: z})
}

type fakeHomer struct {
	calls []string
	err   map[string]error
}

func (f *fakeHomer) DriveTo(_ context.Context, name string) error {
	f.calls = append(f.calls, name)
	if e, ok := f.err[name]; ok {
		return e
	}
	return nil
}

func TestVerifyDrivesBothArms(t *testing.T) {
	m := NewFake()
	err := Verify(context.Background(), []string{"arm-1", "arm-2"}, m, poseAt(400, 0, 300), Options{})
	test.That(t, err, test.ShouldBeNil)
	test.That(t, m.Calls, test.ShouldResemble, []string{"arm-1", "arm-2"})
}

func TestVerifyPropagatesMoveError(t *testing.T) {
	m := NewFake()
	boom := errors.New("unreachable")
	m.MoveError["arm-1"] = boom

	err := Verify(context.Background(), []string{"arm-1", "arm-2"}, m, poseAt(400, 0, 300), Options{})
	test.That(t, err, test.ShouldNotBeNil)
	test.That(t, errors.Is(err, boom), test.ShouldBeTrue)
}

func TestVerifyRejectsWrongArmCount(t *testing.T) {
	err := Verify(context.Background(), []string{"arm-1"}, NewFake(), poseAt(0, 0, 0), Options{})
	test.That(t, err, test.ShouldNotBeNil)
	test.That(t, err.Error(), test.ShouldContainSubstring, "exactly two arms")
}

func TestVerifyRetractsToHomeBetweenArms(t *testing.T) {
	m := NewFake()
	h := &fakeHomer{err: map[string]error{}}

	err := Verify(context.Background(), []string{"arm-1", "arm-2"}, m, poseAt(400, 0, 300), Options{
		Homer:           h,
		HomeSwitchByArm: map[string]string{"arm-1": "arm_1_home", "arm-2": "arm_2_home"},
	})
	test.That(t, err, test.ShouldBeNil)
	test.That(t, h.calls, test.ShouldResemble, []string{"arm_1_home", "arm_2_home"})
}

func TestVerifyDwellRespectsContext(t *testing.T) {
	m := NewFake()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := Verify(ctx, []string{"arm-1", "arm-2"}, m, poseAt(400, 0, 300), Options{Dwell: 10 * time.Second})
	test.That(t, err, test.ShouldEqual, context.Canceled)
}

func TestVerifyRejectsMissingHomeSwitch(t *testing.T) {
	m := NewFake()
	h := &fakeHomer{err: map[string]error{}}

	err := Verify(context.Background(), []string{"arm-1", "arm-2"}, m, poseAt(400, 0, 300), Options{
		Homer:           h,
		HomeSwitchByArm: map[string]string{"arm-1": "arm_1_home"},
	})
	test.That(t, err, test.ShouldNotBeNil)
	test.That(t, err.Error(), test.ShouldContainSubstring, `arm-2`)
}

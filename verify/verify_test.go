package verify

import (
	"testing"

	"go.viam.com/test"
)

func validConfig() *Config {
	return &Config{
		Arms: []string{"arm-1", "arm-2"},
		HomePositions: map[string]string{
			"arm-1": "arm_1_home",
			"arm-2": "arm_2_home",
		},
		VerifyPoses: []VerifyPose{
			{Name: "center", X: 400, Y: 0, Z: 300, OZ: -1},
			{Name: "far_x", X: 500, Y: 0, Z: 300, OZ: -1},
		},
	}
}

func TestValidateHappyPath(t *testing.T) {
	deps, _, err := validConfig().Validate("verify")
	test.That(t, err, test.ShouldBeNil)
	test.That(t, deps, test.ShouldContain, "arm-1")
	test.That(t, deps, test.ShouldContain, "arm-2")
	test.That(t, deps, test.ShouldContain, "arm_1_home")
	test.That(t, deps, test.ShouldContain, "arm_2_home")
	test.That(t, len(deps), test.ShouldEqual, 5)
}

func TestValidateDefaultsMotionToBuiltin(t *testing.T) {
	test.That(t, validConfig().motionName(), test.ShouldEqual, "builtin")
}

func TestValidateHonorsCustomMotionService(t *testing.T) {
	cfg := validConfig()
	cfg.Motion = "custom_motion"
	test.That(t, cfg.motionName(), test.ShouldEqual, "custom_motion")
}

func TestValidateRejectsWrongArmCount(t *testing.T) {
	cfg := validConfig()
	cfg.Arms = []string{"arm-1"}
	_, _, err := cfg.Validate("verify")
	test.That(t, err, test.ShouldNotBeNil)
	test.That(t, err.Error(), test.ShouldContainSubstring, "exactly two arms")
}

func TestValidateRejectsEmptyArmName(t *testing.T) {
	cfg := validConfig()
	cfg.Arms = []string{"arm-1", ""}
	_, _, err := cfg.Validate("verify")
	test.That(t, err, test.ShouldNotBeNil)
	test.That(t, err.Error(), test.ShouldContainSubstring, "arms[1]")
}

func TestValidateRejectsNoVerifyPoses(t *testing.T) {
	cfg := validConfig()
	cfg.VerifyPoses = nil
	_, _, err := cfg.Validate("verify")
	test.That(t, err, test.ShouldNotBeNil)
	test.That(t, err.Error(), test.ShouldContainSubstring, "at least one verify_pose")
}

func TestValidateRejectsMissingPoseName(t *testing.T) {
	cfg := validConfig()
	cfg.VerifyPoses[1].Name = ""
	_, _, err := cfg.Validate("verify")
	test.That(t, err, test.ShouldNotBeNil)
	test.That(t, err.Error(), test.ShouldContainSubstring, "verify_poses[1].name")
}

func TestValidateRejectsDuplicatePoseName(t *testing.T) {
	cfg := validConfig()
	cfg.VerifyPoses[1].Name = "center"
	_, _, err := cfg.Validate("verify")
	test.That(t, err, test.ShouldNotBeNil)
	test.That(t, err.Error(), test.ShouldContainSubstring, "unique names")
}

func TestValidateRejectsMissingHomePosition(t *testing.T) {
	cfg := validConfig()
	delete(cfg.HomePositions, "arm-2")
	_, _, err := cfg.Validate("verify")
	test.That(t, err, test.ShouldNotBeNil)
	test.That(t, err.Error(), test.ShouldContainSubstring, `arm "arm-2"`)
}

func TestValidateRejectsNegativeDwell(t *testing.T) {
	cfg := validConfig()
	cfg.DwellSeconds = -1
	_, _, err := cfg.Validate("verify")
	test.That(t, err, test.ShouldNotBeNil)
	test.That(t, err.Error(), test.ShouldContainSubstring, "non-negative")
}

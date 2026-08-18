package touch

import (
	"testing"

	"go.viam.com/test"
)

func validConfig() *Config {
	return &Config{
		Arms: []ArmConfig{
			{
				Name:           "arm_a",
				HomePosition:   "arm_a_home",
				PositionSavers: []string{"arm_a_p1", "arm_a_p2", "arm_a_p3", "arm_a_p4"},
			},
			{
				Name:           "arm_b",
				HomePosition:   "arm_b_home",
				PositionSavers: []string{"arm_b_p1", "arm_b_p2", "arm_b_p3", "arm_b_p4"},
			},
		},
	}
}

func TestValidateHappyPath(t *testing.T) {
	cfg := validConfig()
	deps, _, err := cfg.Validate("touch")
	test.That(t, err, test.ShouldBeNil)
	test.That(t, deps, test.ShouldContain, "arm_a")
	test.That(t, deps, test.ShouldContain, "arm_b")
	test.That(t, deps, test.ShouldContain, "arm_a_home")
	test.That(t, deps, test.ShouldContain, "arm_b_home")
	test.That(t, deps, test.ShouldContain, "arm_a_p1")
	test.That(t, deps, test.ShouldContain, "arm_b_p4")
	test.That(t, len(deps), test.ShouldEqual, 12)
}

func TestValidateRejectsFewerThanTwoArms(t *testing.T) {
	cfg := validConfig()
	cfg.Arms = cfg.Arms[:1]
	_, _, err := cfg.Validate("touch")
	test.That(t, err, test.ShouldNotBeNil)
	test.That(t, err.Error(), test.ShouldContainSubstring, "exactly two arms")
}

func TestValidateRejectsMoreThanTwoArms(t *testing.T) {
	cfg := validConfig()
	cfg.Arms = append(cfg.Arms, ArmConfig{
		Name: "arm_c", HomePosition: "arm_c_home",
		PositionSavers: []string{"c1", "c2", "c3", "c4"},
	})
	_, _, err := cfg.Validate("touch")
	test.That(t, err, test.ShouldNotBeNil)
	test.That(t, err.Error(), test.ShouldContainSubstring, "exactly two arms")
}

func TestValidateRejectsMissingArmName(t *testing.T) {
	cfg := validConfig()
	cfg.Arms[1].Name = ""
	_, _, err := cfg.Validate("touch")
	test.That(t, err, test.ShouldNotBeNil)
	test.That(t, err.Error(), test.ShouldContainSubstring, "arms[1].name")
}

func TestValidateRejectsMissingHomePosition(t *testing.T) {
	cfg := validConfig()
	cfg.Arms[0].HomePosition = ""
	_, _, err := cfg.Validate("touch")
	test.That(t, err, test.ShouldNotBeNil)
	test.That(t, err.Error(), test.ShouldContainSubstring, "arms[0].home_position")
}

func TestValidateRejectsFewerThanFourPositionSavers(t *testing.T) {
	cfg := validConfig()
	cfg.Arms[0].PositionSavers = cfg.Arms[0].PositionSavers[:3]
	cfg.Arms[1].PositionSavers = cfg.Arms[1].PositionSavers[:3]
	_, _, err := cfg.Validate("touch")
	test.That(t, err, test.ShouldNotBeNil)
	test.That(t, err.Error(), test.ShouldContainSubstring, "at least four position_savers")
}

func TestValidateRejectsMismatchedPositionSaverCounts(t *testing.T) {
	cfg := validConfig()
	cfg.Arms[1].PositionSavers = append(cfg.Arms[1].PositionSavers, "arm_b_p5")
	_, _, err := cfg.Validate("touch")
	test.That(t, err, test.ShouldNotBeNil)
	test.That(t, err.Error(), test.ShouldContainSubstring, "same number of position_savers")
}

func TestValidateRejectsEmptyPositionSaverEntry(t *testing.T) {
	cfg := validConfig()
	cfg.Arms[0].PositionSavers[2] = ""
	_, _, err := cfg.Validate("touch")
	test.That(t, err, test.ShouldNotBeNil)
	test.That(t, err.Error(), test.ShouldContainSubstring, "arms[0].position_savers[2]")
}

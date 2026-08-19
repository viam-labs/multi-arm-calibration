// Package verify implements the multi-arm calibration verify service:
// plan both arms to a set of target world poses declared in config, dwell for
// human observation, retract each arm to home between them, and report
// per-pose per-arm miss + inter-arm separation plus aggregate stats.
package verify

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/geo/r3"
	toggleswitch "go.viam.com/rdk/components/switch"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/services/generic"
	"go.viam.com/rdk/services/motion"
	"go.viam.com/rdk/spatialmath"

	"github.com/viam-labs/multi-arm-calibration/internal/posdriver"
	"github.com/viam-labs/multi-arm-calibration/internal/verifier"
)

var Model = resource.NewModel("viam", "multi-arm-calibration", "verify")

func init() {
	resource.RegisterService(generic.API, Model,
		resource.Registration[resource.Resource, *Config]{
			Constructor: newVerify,
		},
	)
}

type Config struct {
	Arms           []string          `json:"arms"`
	Motion         string            `json:"motion,omitempty"`
	HomePositions  map[string]string `json:"home_positions"`
	VerifyPoses    []VerifyPose      `json:"verify_poses"`
	DwellSeconds   float64           `json:"dwell_seconds,omitempty"`
}

type VerifyPose struct {
	Name  string  `json:"name"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Z     float64 `json:"z"`
	OX    float64 `json:"o_x"`
	OY    float64 `json:"o_y"`
	OZ    float64 `json:"o_z"`
	Theta float64 `json:"theta"`
}

func (p VerifyPose) pose() spatialmath.Pose {
	return spatialmath.NewPose(
		r3.Vector{X: p.X, Y: p.Y, Z: p.Z},
		&spatialmath.OrientationVectorDegrees{OX: p.OX, OY: p.OY, OZ: p.OZ, Theta: p.Theta},
	)
}

func (cfg *Config) motionName() string {
	if cfg.Motion == "" {
		return "builtin"
	}
	return cfg.Motion
}

func (cfg *Config) Validate(path string) ([]string, []string, error) {
	if len(cfg.Arms) != 2 {
		return nil, nil, resource.NewConfigValidationError(path, errExactlyTwoArms)
	}
	if len(cfg.VerifyPoses) < 1 {
		return nil, nil, resource.NewConfigValidationError(path, errAtLeastOneVerifyPose)
	}
	names := make(map[string]struct{}, len(cfg.VerifyPoses))
	for i, p := range cfg.VerifyPoses {
		if p.Name == "" {
			return nil, nil, resource.NewConfigValidationFieldRequiredError(path, fmt.Sprintf("verify_poses[%d].name", i))
		}
		if _, dup := names[p.Name]; dup {
			return nil, nil, resource.NewConfigValidationError(path, errDuplicateVerifyPose)
		}
		names[p.Name] = struct{}{}
	}
	if cfg.DwellSeconds < 0 {
		return nil, nil, resource.NewConfigValidationError(path, errNegativeDwell)
	}

	deps := make([]string, 0, 2+len(cfg.HomePositions)+1)
	for i, name := range cfg.Arms {
		if name == "" {
			return nil, nil, resource.NewConfigValidationFieldRequiredError(path, fmt.Sprintf("arms[%d]", i))
		}
		deps = append(deps, name)
		home, ok := cfg.HomePositions[name]
		if !ok || home == "" {
			return nil, nil, resource.NewConfigValidationError(path, fmt.Errorf("home_positions must include arm %q", name))
		}
		deps = append(deps, home)
	}
	deps = append(deps, motion.Named(cfg.motionName()).String())
	return deps, nil, nil
}

type service struct {
	resource.AlwaysRebuild
	resource.Named
	logger        logging.Logger
	arms          []string
	motion        motion.Service
	poses         []VerifyPose
	homeSwitches  map[string]string
	driver        *posdriver.Switches
	dwell         time.Duration
}

func newVerify(_ context.Context, deps resource.Dependencies, conf resource.Config, logger logging.Logger) (resource.Resource, error) {
	cfg, err := resource.NativeConfig[*Config](conf)
	if err != nil {
		return nil, err
	}
	ms, err := motion.FromProvider(deps, cfg.motionName())
	if err != nil {
		return nil, fmt.Errorf("resolve motion service %q: %w", cfg.motionName(), err)
	}

	switches := make(map[string]toggleswitch.Switch, len(cfg.HomePositions))
	for _, armName := range cfg.Arms {
		swName := cfg.HomePositions[armName]
		sw, err := toggleswitch.FromProvider(deps, swName)
		if err != nil {
			return nil, fmt.Errorf("resolve home switch %q for arm %q: %w", swName, armName, err)
		}
		switches[swName] = sw
	}

	return &service{
		Named:        conf.ResourceName().AsNamed(),
		logger:       logger,
		arms:         cfg.Arms,
		motion:       ms,
		poses:        cfg.VerifyPoses,
		homeSwitches: cfg.HomePositions,
		driver:       posdriver.NewSwitches(switches),
		dwell:        time.Duration(cfg.DwellSeconds * float64(time.Second)),
	}, nil
}

func (s *service) DoCommand(ctx context.Context, _ map[string]interface{}) (map[string]interface{}, error) {
	opts := verifier.Options{
		HomeSwitchByArm: s.homeSwitches,
		Homer:           s.driver,
		Dwell:           s.dwell,
	}

	var failed []interface{}
	for _, p := range s.poses {
		if err := verifier.Verify(ctx, s.arms, s.motion, p.pose(), opts); err != nil {
			s.logger.Warnf("pose %q failed: %v", p.Name, err)
			failed = append(failed, p.Name)
		}
	}
	return map[string]interface{}{
		"success":      len(failed) == 0,
		"failed_poses": failed,
	}, nil
}

func (s *service) Close(_ context.Context) error {
	return nil
}

// Package touch implements touch-based multi-arm base calibration.
package touch

import (
	"context"
	"fmt"

	"go.viam.com/rdk/components/arm"
	toggleswitch "go.viam.com/rdk/components/switch"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/services/generic"
	"go.viam.com/rdk/spatialmath"

	"github.com/viam-labs/multi-arm-calibration/internal/posdriver"
	"github.com/viam-labs/multi-arm-calibration/internal/posesource"
)

var Model = resource.NewModel("viam", "multi-arm-calibration", "touch")

func init() {
	resource.RegisterService(generic.API, Model,
		resource.Registration[resource.Resource, *Config]{
			Constructor: newTouch,
		},
	)
}

type Config struct {
	Arms []ArmConfig `json:"arms"`
}

type ArmConfig struct {
	Name           string   `json:"name"`
	HomePosition   string   `json:"home_position"`
	PositionSavers []string `json:"position_savers"`
}

func (cfg *Config) Validate(path string) ([]string, []string, error) {
	if len(cfg.Arms) != 2 {
		return nil, nil, resource.NewConfigValidationError(path, errExactlyTwoArms)
	}

	expectedSavers := len(cfg.Arms[0].PositionSavers)
	if expectedSavers < 4 {
		return nil, nil, resource.NewConfigValidationError(path, errAtLeastFourPositionSavers)
	}

	deps := make([]string, 0, 2+2+2*expectedSavers)
	for i, a := range cfg.Arms {
		if a.Name == "" {
			return nil, nil, resource.NewConfigValidationFieldRequiredError(path, fmt.Sprintf("arms[%d].name", i))
		}
		if a.HomePosition == "" {
			return nil, nil, resource.NewConfigValidationFieldRequiredError(path, fmt.Sprintf("arms[%d].home_position", i))
		}
		if len(a.PositionSavers) != expectedSavers {
			return nil, nil, resource.NewConfigValidationError(path, errMismatchedPositionSaverCounts)
		}
		for j, s := range a.PositionSavers {
			if s == "" {
				return nil, nil, resource.NewConfigValidationFieldRequiredError(path, fmt.Sprintf("arms[%d].position_savers[%d]", i, j))
			}
			deps = append(deps, s)
		}
		deps = append(deps, a.Name, a.HomePosition)
	}
	return deps, nil, nil
}

type service struct {
	resource.AlwaysRebuild
	resource.Named
	logger logging.Logger
	cfg    *Config
	poser  *posesource.Arms
	driver *posdriver.Switches
}

func newTouch(_ context.Context, deps resource.Dependencies, conf resource.Config, logger logging.Logger) (resource.Resource, error) {
	cfg, err := resource.NativeConfig[*Config](conf)
	if err != nil {
		return nil, err
	}

	arms := make(map[string]arm.Arm, len(cfg.Arms))
	switches := make(map[string]toggleswitch.Switch)
	for _, a := range cfg.Arms {
		armDep, err := arm.FromProvider(deps, a.Name)
		if err != nil {
			return nil, fmt.Errorf("resolve arm %q: %w", a.Name, err)
		}
		arms[a.Name] = armDep

		for _, sw := range append([]string{a.HomePosition}, a.PositionSavers...) {
			if _, seen := switches[sw]; seen {
				continue
			}
			swDep, err := toggleswitch.FromProvider(deps, sw)
			if err != nil {
				return nil, fmt.Errorf("resolve switch %q: %w", sw, err)
			}
			switches[sw] = swDep
		}
	}

	return &service{
		Named:  conf.ResourceName().AsNamed(),
		logger: logger,
		cfg:    cfg,
		poser:  posesource.NewArms(arms),
		driver: posdriver.NewSwitches(switches),
	}, nil
}

func (s *service) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	if _, ok := cmd["calibrate"]; ok {
		t, err := Calibrate(ctx, s.cfg, s.driver, s.poser)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"transform": poseToMap(t)}, nil
	}
	return nil, fmt.Errorf("unknown command: %v", cmd)
}

func poseToMap(p spatialmath.Pose) map[string]interface{} {
	t := p.Point()
	ov := p.Orientation().OrientationVectorDegrees()
	return map[string]interface{}{
		"translation_mm":     map[string]interface{}{"x": t.X, "y": t.Y, "z": t.Z},
		"orientation_ov_deg": map[string]interface{}{"ox": ov.OX, "oy": ov.OY, "oz": ov.OZ, "theta": ov.Theta},
	}
}

func (s *service) Close(_ context.Context) error {
	return nil
}

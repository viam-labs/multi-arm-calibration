// Package touch implements touch-based multi-arm base calibration.
package touch

import (
	"context"
	"fmt"

	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/services/generic"
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
}

func newTouch(_ context.Context, _ resource.Dependencies, conf resource.Config, logger logging.Logger) (resource.Resource, error) {
	return &service{
		Named:  conf.ResourceName().AsNamed(),
		logger: logger,
	}, nil
}

func (s *service) DoCommand(_ context.Context, _ map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{"status": "not implemented"}, nil
}

func (s *service) Close(_ context.Context) error {
	return nil
}

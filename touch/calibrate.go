package touch

import (
	"context"
	"fmt"

	"github.com/golang/geo/r3"
	"go.viam.com/rdk/spatialmath"

	"github.com/viam-labs/multi-arm-calibration/internal/solver"
)

type driver interface {
	DriveTo(ctx context.Context, switchName string) error
}

type poser interface {
	Capture(ctx context.Context, armName string) (spatialmath.Pose, error)
}

// Calibrate drives both arms through their saved touch points, captures each
// arm's TCP in its own base frame, and returns the rigid transform T such that
// T · P_armA ≈ P_armB for the shared touch points.
func Calibrate(ctx context.Context, cfg *Config, d driver, p poser) (spatialmath.Pose, error) {
	if len(cfg.Arms) != 2 {
		return nil, fmt.Errorf("calibrate needs exactly two arms, got %d", len(cfg.Arms))
	}

	points := make(map[string][]r3.Vector, len(cfg.Arms))
	for _, a := range cfg.Arms {
		pts, err := captureArm(ctx, a, d, p)
		if err != nil {
			return nil, err
		}
		points[a.Name] = pts
	}

	return solver.Solve(points[cfg.Arms[0].Name], points[cfg.Arms[1].Name])
}

func captureArm(ctx context.Context, a ArmConfig, d driver, p poser) ([]r3.Vector, error) {
	if err := d.DriveTo(ctx, a.HomePosition); err != nil {
		return nil, fmt.Errorf("arm %q: drive to home: %w", a.Name, err)
	}

	pts := make([]r3.Vector, 0, len(a.PositionSavers))
	for i, saver := range a.PositionSavers {
		if err := d.DriveTo(ctx, saver); err != nil {
			return nil, fmt.Errorf("arm %q: drive to saver[%d]=%q: %w", a.Name, i, saver, err)
		}
		pose, err := p.Capture(ctx, a.Name)
		if err != nil {
			return nil, fmt.Errorf("arm %q: capture at saver[%d]=%q: %w", a.Name, i, saver, err)
		}
		pts = append(pts, pose.Point())
	}

	if err := d.DriveTo(ctx, a.HomePosition); err != nil {
		return nil, fmt.Errorf("arm %q: return to home: %w", a.Name, err)
	}
	return pts, nil
}

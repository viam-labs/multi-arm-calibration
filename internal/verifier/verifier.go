// Package verifier drives both arms sequentially to a target world pose,
// dwelling for human observation and retracting each arm to a safe home
// between them. It measures nothing on its own — physical alignment is
// judged by the operator watching the arms land.
package verifier

import (
	"context"
	"fmt"
	"time"

	"go.viam.com/rdk/referenceframe"
	"go.viam.com/rdk/services/motion"
	"go.viam.com/rdk/spatialmath"
)

type mover interface {
	Move(ctx context.Context, req motion.MoveReq) (bool, error)
}

type homer interface {
	DriveTo(ctx context.Context, switchName string) error
}

type Options struct {
	HomeSwitchByArm map[string]string
	Homer           homer
	Dwell           time.Duration
}

func Verify(ctx context.Context, arms []string, m mover, target spatialmath.Pose, opts Options) error {
	if len(arms) != 2 {
		return fmt.Errorf("verify needs exactly two arms, got %d", len(arms))
	}

	dest := referenceframe.NewPoseInFrame(referenceframe.World, target)
	for _, armName := range arms {
		if _, err := m.Move(ctx, motion.MoveReq{ComponentName: armName, Destination: dest}); err != nil {
			return fmt.Errorf("arm %q: move: %w", armName, err)
		}

		if opts.Dwell > 0 {
			select {
			case <-time.After(opts.Dwell):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		if opts.Homer != nil {
			home, ok := opts.HomeSwitchByArm[armName]
			if !ok {
				return fmt.Errorf("arm %q: no home switch configured", armName)
			}
			if err := opts.Homer.DriveTo(ctx, home); err != nil {
				return fmt.Errorf("arm %q: retract to home: %w", armName, err)
			}
		}
	}

	return nil
}

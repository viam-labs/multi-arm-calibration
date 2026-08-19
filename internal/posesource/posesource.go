// Package posesource captures TCP poses from named arms.
package posesource

import (
	"context"
	"fmt"

	"go.viam.com/rdk/components/arm"
	"go.viam.com/rdk/spatialmath"
)

type Arms struct {
	arms map[string]arm.Arm
}

func NewArms(arms map[string]arm.Arm) *Arms {
	return &Arms{arms: arms}
}

// Capture returns the arm's TCP pose in its own base frame.
func (a *Arms) Capture(ctx context.Context, armName string) (spatialmath.Pose, error) {
	arm, ok := a.arms[armName]
	if !ok {
		return nil, fmt.Errorf("unknown arm %q", armName)
	}
	return arm.EndPosition(ctx, nil)
}

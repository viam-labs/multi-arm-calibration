// Package posdriver drives arms to saved poses via arm-position-saver switches.
package posdriver

import (
	"context"
	"fmt"

	toggleswitch "go.viam.com/rdk/components/switch"
)

// goToPosition is the arm-position-saver switch position that drives the arm
// to its saved pose. Position 1 rewrites the config; we never want that.
const goToPosition uint32 = 2

type Switches struct {
	switches map[string]toggleswitch.Switch
}

func NewSwitches(switches map[string]toggleswitch.Switch) *Switches {
	return &Switches{switches: switches}
}

// DriveTo toggles the named switch to "go to" and blocks until the arm arrives.
func (s *Switches) DriveTo(ctx context.Context, switchName string) error {
	sw, ok := s.switches[switchName]
	if !ok {
		return fmt.Errorf("unknown switch %q", switchName)
	}
	return sw.SetPosition(ctx, goToPosition, nil)
}

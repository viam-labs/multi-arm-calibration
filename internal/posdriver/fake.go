package posdriver

import (
	"context"
	"fmt"
)

// Fake records the order in which switches were driven and can inject errors per switch for tests.
type Fake struct {
	Known  map[string]struct{}
	Errors map[string]error
	Calls  []string
}

func NewFake(known ...string) *Fake {
	set := make(map[string]struct{}, len(known))
	for _, k := range known {
		set[k] = struct{}{}
	}
	return &Fake{Known: set, Errors: map[string]error{}}
}

func (f *Fake) DriveTo(_ context.Context, switchName string) error {
	if _, ok := f.Known[switchName]; !ok {
		return fmt.Errorf("unknown switch %q", switchName)
	}
	f.Calls = append(f.Calls, switchName)
	if err, ok := f.Errors[switchName]; ok {
		return err
	}
	return nil
}

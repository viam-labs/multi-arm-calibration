package posesource

import (
	"context"
	"fmt"

	"go.viam.com/rdk/spatialmath"
)

// Fake serves canned pose sequences per arm name for tests.
// Each Capture call for a given arm returns the next pose in that arm's queue.
type Fake struct {
	poses map[string][]spatialmath.Pose
	next  map[string]int
}

func NewFake(poses map[string][]spatialmath.Pose) *Fake {
	return &Fake{
		poses: poses,
		next:  make(map[string]int),
	}
}

func (f *Fake) Capture(_ context.Context, armName string) (spatialmath.Pose, error) {
	q, ok := f.poses[armName]
	if !ok {
		return nil, fmt.Errorf("unknown arm %q", armName)
	}
	i := f.next[armName]
	if i >= len(q) {
		return nil, fmt.Errorf("arm %q exhausted after %d captures", armName, len(q))
	}
	f.next[armName] = i + 1
	return q[i], nil
}

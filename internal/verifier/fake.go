package verifier

import (
	"context"

	"go.viam.com/rdk/services/motion"
)

type Fake struct {
	MoveError map[string]error
	Calls     []string
}

func NewFake() *Fake {
	return &Fake{MoveError: map[string]error{}}
}

func (f *Fake) Move(_ context.Context, req motion.MoveReq) (bool, error) {
	f.Calls = append(f.Calls, req.ComponentName)
	if err, ok := f.MoveError[req.ComponentName]; ok {
		return false, err
	}
	return true, nil
}

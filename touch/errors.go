package touch

import "errors"

var (
	errExactlyTwoArms                = errors.New("must declare exactly two arms")
	errAtLeastFourPositionSavers     = errors.New("each arm must declare at least four position_savers")
	errMismatchedPositionSaverCounts = errors.New("all arms must declare the same number of position_savers")
)

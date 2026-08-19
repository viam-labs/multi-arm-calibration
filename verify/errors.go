package verify

import "errors"

var (
	errExactlyTwoArms       = errors.New("must declare exactly two arms")
	errAtLeastOneVerifyPose = errors.New("must declare at least one verify_pose")
	errDuplicateVerifyPose  = errors.New("verify_poses must have unique names")
	errNegativeDwell        = errors.New("dwell_seconds must be non-negative")
)

package posdriver

import (
	"context"
	"errors"
	"testing"

	"go.viam.com/test"
)

func TestFakeRecordsCallOrder(t *testing.T) {
	f := NewFake("a_home", "a_p1", "a_p2")

	test.That(t, f.DriveTo(context.Background(), "a_home"), test.ShouldBeNil)
	test.That(t, f.DriveTo(context.Background(), "a_p1"), test.ShouldBeNil)
	test.That(t, f.DriveTo(context.Background(), "a_p2"), test.ShouldBeNil)
	test.That(t, f.DriveTo(context.Background(), "a_home"), test.ShouldBeNil)

	test.That(t, f.Calls, test.ShouldResemble, []string{"a_home", "a_p1", "a_p2", "a_home"})
}

func TestFakeRejectsUnknownSwitch(t *testing.T) {
	f := NewFake("a_home")
	err := f.DriveTo(context.Background(), "missing")
	test.That(t, err, test.ShouldNotBeNil)
	test.That(t, err.Error(), test.ShouldContainSubstring, `unknown switch "missing"`)
}

func TestFakePropagatesConfiguredError(t *testing.T) {
	f := NewFake("a_p1")
	boom := errors.New("boom")
	f.Errors["a_p1"] = boom

	err := f.DriveTo(context.Background(), "a_p1")
	test.That(t, err, test.ShouldEqual, boom)
	test.That(t, f.Calls, test.ShouldResemble, []string{"a_p1"})
}

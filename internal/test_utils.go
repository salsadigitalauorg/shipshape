package internal

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
)

func EnsureFail(t *testing.T, c *shipshape.CheckBase) (msg string, ok bool) {
	if c.Result.Status != shipshape.Fail {
		return fmt.Sprintf("Check should Fail, got %#v", c.Result.Passes), false
	}
	return "", true
}

// EnsureNoFail is different from EnsurePass because Pass is the final result
// while during various steps the Status might not yet be determined, but should
// still not fail.
func EnsureNoFail(t *testing.T, c *shipshape.CheckBase) (msg string, ok bool) {
	if c.Result.Status == shipshape.Fail {
		return fmt.Sprintf("Check should not Fail yet, got %#v", c.Result.Failures), false
	}
	return "", true
}

func EnsurePass(t *testing.T, c *shipshape.CheckBase) (msg string, ok bool) {
	if c.Result.Status != shipshape.Pass {
		return fmt.Sprintf("Check should Pass, got %#v", c.Result.Failures), false
	}
	return "", true
}

func EnsureFailures(t *testing.T, c *shipshape.CheckBase, expectedFailures []string) (msg string, ok bool) {
	numExpectedFailures := len(expectedFailures)
	if len(c.Result.Failures) != numExpectedFailures || !reflect.DeepEqual(expectedFailures, c.Result.Failures) {
		return fmt.Sprintf("there should be exactly %d Failure(s), got %#v", numExpectedFailures, c.Result.Failures), false
	}
	return "", true
}

func EnsurePasses(t *testing.T, c *shipshape.CheckBase, expectedPasses []string) (msg string, ok bool) {
	numExpectedPasses := len(expectedPasses)
	if len(c.Result.Passes) != numExpectedPasses || !reflect.DeepEqual(expectedPasses, c.Result.Passes) {
		return fmt.Sprintf("there should be exactly %d Pass(es), got %#v", numExpectedPasses, c.Result.Passes), false
	}
	return "", true
}

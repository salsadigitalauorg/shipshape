package internal

import (
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
)

var FakeCommandArgs []string
var MockedExitStatus = 0
var MockedStdout string
var MockedStderr string

func FakeExecCommand(command string, args ...string) *exec.Cmd {
	FakeCommandArgs = args
	cs := []string{"-test.run=TestExecCommandHelper", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	es := strconv.Itoa(MockedExitStatus)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1",
		"STDOUT=" + MockedStdout,
		"STDERR=" + MockedStderr,
		"EXIT_STATUS=" + es}
	return cmd
}

// TestExecCommandHelper prints the result of a fake command to either
// stderr or stdout, depending on the exit status.
// This func should not be called directly, but instead should be called from
// a new definition inside the test file where fake commands are tested.
// For example, create the following at the top of the file:
// 		func TestExecCommandHelper(t *testing.T) {
//			internal.TestExecCommandHelper(t)
// 		}
func TestExecCommandHelper(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	i, _ := strconv.Atoi(os.Getenv("EXIT_STATUS"))
	if i > 0 {
		fmt.Fprint(os.Stderr, os.Getenv("STDERR"))
	} else {
		fmt.Fprint(os.Stdout, os.Getenv("STDOUT"))
	}
	os.Exit(i)
}

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

func EnsureWarnings(t *testing.T, c *shipshape.CheckBase, expectedWarnings []string) (msg string, ok bool) {
	numExpectedWarnings := len(expectedWarnings)
	if len(c.Result.Warnings) != numExpectedWarnings || !reflect.DeepEqual(expectedWarnings, c.Result.Warnings) {
		return fmt.Sprintf("there should be exactly %d Warning(s), got %#v", numExpectedWarnings, c.Result.Warnings), false
	}
	return "", true
}

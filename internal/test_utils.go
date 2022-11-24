package internal

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"testing"
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
//
//	func TestExecCommandHelper(t *testing.T) {
//		internal.TestExecCommandHelper(t)
//	}
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

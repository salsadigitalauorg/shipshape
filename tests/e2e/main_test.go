package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestMain builds the shipshape binary once and shares it across every test in
// the package. Building the real CLI (rather than calling packages in-process)
// keeps these tests genuinely end-to-end: flag parsing, output rendering and
// process exit codes are all exercised exactly as a user would hit them.
func TestMain(m *testing.M) {
	code, err := run(m)
	if err != nil {
		fmt.Fprintln(os.Stderr, "e2e setup failed:", err)
		os.Exit(1)
	}
	os.Exit(code)
}

func run(m *testing.M) (int, error) {
	// Ryuk (the Testcontainers reaper) bind-mounts the Docker socket, which
	// fails under some Docker hosts (notably Colima, where the socket is not a
	// bind-mountable path). Every container-backed test registers an explicit
	// t.Cleanup that terminates its container, so disabling the reaper is safe
	// and makes the suite portable across Docker Desktop, Colima and CI. Can be
	// overridden by exporting TESTCONTAINERS_RYUK_DISABLED before running.
	if _, set := os.LookupEnv("TESTCONTAINERS_RYUK_DISABLED"); !set {
		os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	}

	root, err := findRepoRoot()
	if err != nil {
		return 0, err
	}
	repoRoot = root

	binDir, err := os.MkdirTemp("", "shipshape-e2e-bin")
	if err != nil {
		return 0, fmt.Errorf("create temp bin dir: %w", err)
	}
	defer os.RemoveAll(binDir)

	bin := filepath.Join(binDir, "shipshape")
	if err := buildBinary(root, bin); err != nil {
		return 0, err
	}
	shipshapeBin = bin

	return m.Run(), nil
}

// buildBinary compiles the shipshape binary from the repository root.
func buildBinary(root, out string) error {
	cmd := exec.Command("go", "build", "-o", out, ".")
	cmd.Dir = root
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build shipshape binary: %w", err)
	}
	return nil
}

// findRepoRoot walks up from the current working directory until it finds a
// go.mod, returning that directory. Tests run with the working directory set to
// the package (tests/e2e), so the module root is two levels up, but walking is
// resilient to that assumption changing.
func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not locate go.mod above %s", dir)
		}
		dir = parent
	}
}

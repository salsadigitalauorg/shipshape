package e2e

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// shipshapeBin is the absolute path to the shipshape binary built once for the
// whole e2e package by TestMain.
var shipshapeBin string

// repoRoot is the absolute path to the repository root.
var repoRoot string

// ResultList mirrors pkg/result.ResultList for decoding shipshape's `-o json`
// output. It is a deliberately independent copy: the real Result.Breaches field
// is a breach.Breach interface which cannot be unmarshalled directly, and
// keeping the e2e assertions decoupled from internal types means an internal
// refactor cannot silently weaken these tests. Breaches are captured as loose
// maps so tests can assert on any breach attribute.
type ResultList struct {
	Policies              map[string][]string `json:"policies"`
	RemediationPerformed  bool                `json:"remediation-performed"`
	TotalChecks           uint32              `json:"total-checks"`
	TotalBreaches         uint32              `json:"total-breaches"`
	CheckCountByType      map[string]int      `json:"check-count-by-type"`
	BreachCountByType     map[string]int      `json:"breach-count-by-type"`
	BreachCountBySeverity map[string]int      `json:"breach-count-by-severity"`
	Results               []Result            `json:"results"`
}

// Result mirrors pkg/result.Result for the fields e2e tests assert on.
type Result struct {
	Name      string           `json:"name"`
	Severity  string           `json:"severity"`
	CheckType string           `json:"check-type"`
	Status    string           `json:"status"`
	Passes    []string         `json:"passes"`
	Breaches  []map[string]any `json:"breaches"`
}

// RunResult captures everything a subprocess invocation of shipshape produced,
// so tests can assert on stdout, stderr and the process exit code independently.
type RunResult struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
}

// runShipshape executes the compiled shipshape binary with the given working
// directory and arguments, and returns the captured result. It never fails the
// test itself: a non-zero exit code from shipshape is a legitimate outcome
// (e.g. with -e when breaches are found), so callers assert on ExitCode.
func runShipshape(t *testing.T, workdir string, args ...string) RunResult {
	t.Helper()

	cmd := exec.Command(shipshapeBin, args...)
	cmd.Dir = workdir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	res := RunResult{
		Stdout: stdout.Bytes(),
		Stderr: stderr.Bytes(),
	}
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			res.ExitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to run shipshape %v: %v\nstderr: %s",
				args, err, stderr.String())
		}
	}
	return res
}

// runExample is the common invocation for an example config: run the binary
// against the given working directory, pointing -f at the example file under
// examples/, and rendering to the requested output format.
func runExample(t *testing.T, workdir, exampleFile, format string, extraArgs ...string) RunResult {
	t.Helper()

	args := []string{"run", ".", "-f", examplePath(exampleFile), "-o", format}
	args = append(args, extraArgs...)
	return runShipshape(t, workdir, args...)
}

// DecodeJSON unmarshals the stdout of a `-o json` run into the test-local
// ResultList mirror. It fails the test with a helpful message (including
// stderr) if stdout is not valid JSON, which is the usual symptom of a fatal
// error inside shipshape.
func (r RunResult) DecodeJSON(t *testing.T) *ResultList {
	t.Helper()

	trimmed := bytes.TrimSpace(r.Stdout)
	if len(trimmed) == 0 {
		t.Fatalf("expected JSON on stdout but it was empty (exit=%d)\nstderr: %s",
			r.ExitCode, r.Stderr)
	}

	var rl ResultList
	if err := json.Unmarshal(trimmed, &rl); err != nil {
		t.Fatalf("failed to decode JSON output: %v\nstdout: %s\nstderr: %s",
			err, r.Stdout, r.Stderr)
	}
	return &rl
}

// examplePath returns the absolute path to a file in the examples/ directory.
func examplePath(name string) string {
	return filepath.Join(repoRoot, "examples", name)
}

// examplesTestdata returns the absolute path to examples/testdata, the source
// of fixtures most self-contained examples expect.
func examplesTestdata() string {
	return filepath.Join(repoRoot, "examples", "testdata")
}

// stagedTestdata copies examples/testdata into a fresh per-test temp directory
// and returns its path. Tests run examples against this copy rather than the
// real examples/testdata so that example side effects (several examples define
// an `output.file` writing results.xml into the working directory) never
// pollute the repository. The copy is cheap — testdata is a handful of small
// YAML files.
func stagedTestdata(t *testing.T) string {
	t.Helper()

	dst := t.TempDir()
	src := examplesTestdata()

	err := filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		t.Fatalf("stage testdata: %v", err)
	}
	return dst
}

// findResult returns the named result and true if present in the list.
func findResult(rl *ResultList, name string) (Result, bool) {
	for _, res := range rl.Results {
		if res.Name == name {
			return res, true
		}
	}
	return Result{}, false
}

// stdoutContains reports whether the combined stdout contains substr.
func (r RunResult) stdoutContains(substr string) bool {
	return strings.Contains(string(r.Stdout), substr)
}

// writeFile writes content to path, creating parent directories, and registers
// cleanup via the test. Used to seed fixture trees under a temp dir.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir for fixture %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture %s: %v", path, err)
	}
}

package output_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	. "github.com/salsadigitalauorg/shipshape/pkg/output"
	"github.com/salsadigitalauorg/shipshape/pkg/remediation"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
)

func TestFileOutput(t *testing.T) {
	tt := []struct {
		name        string
		file        *File
		rl          *result.ResultList
		expected    string
		expectError bool
		expectNil   bool
	}{
		{
			name: "noResult",
			file: &File{
				Path:   "testdata/pretty.txt",
				Format: "pretty",
			},
			rl:          &result.ResultList{},
			expected:    "No result available; ensure your shipshape.yml is configured correctly.\n",
			expectError: false,
			expectNil:   false,
		},
		{
			name: "pretty format with passes",
			file: &File{
				Path:   "testdata/pretty.txt",
				Format: "pretty",
			},
			rl: &result.ResultList{
				Results: []result.Result{
					{
						Name:   "test-check",
						Status: result.Pass,
					},
				},
			},
			expected:    "Ship is in top shape; no breach detected!\n",
			expectError: false,
			expectNil:   false,
		},
		{
			name: "pretty format with breaches",
			file: &File{
				Path:   "testdata/pretty.txt",
				Format: "pretty",
			},
			rl: &result.ResultList{
				Results: []result.Result{
					{
						Name:            "test-check",
						Status:          result.Fail,
						DetailedMessage: "Custom message for test-check",
						Breaches: []breach.Breach{
							&breach.ValueBreach{Value: "Fail b"},
						},
					},
				},
			},
			expected:    "# Breaches were detected\n\n  ### test-check\n     -- Fail b\n     #- Custom message for test-check\n\n",
			expectError: false,
			expectNil:   false,
		},
		{
			name: "table format with mixed results",
			file: &File{
				Path:   "testdata/table.txt",
				Format: "table",
			},
			rl: &result.ResultList{
				Results: []result.Result{
					{
						Name:   "a",
						Status: result.Pass,
						Passes: []string{"Pass a", "Pass ab"},
					},
					{
						Name:   "b",
						Status: result.Fail,
						Breaches: []breach.Breach{
							&breach.ValueBreach{Value: "Fail b"},
						},
					},
				},
			},
			expected: "NAME   STATUS   PASSES    FAILS\n" +
				"a      Pass     Pass a    \n" +
				"                Pass ab   \n" +
				"b      Fail               Fail b\n",
			expectError: false,
			expectNil:   false,
		},
		{
			name: "json format with remediation",
			file: &File{
				Path:   "testdata/results.json",
				Format: "json",
			},
			rl: &result.ResultList{
				Results: []result.Result{
					{
						Name:   "test-check",
						Status: result.Fail,
						Breaches: []breach.Breach{
							&breach.ValueBreach{
								Value: "Fail b",
								RemediationResult: remediation.RemediationResult{
									Status:   remediation.RemediationStatusSuccess,
									Messages: []string{"fixed 1"},
								},
							},
						},
					},
				},
				RemediationPerformed: true,
				RemediationTotals:    map[string]uint32{"successful": 1},
			},
			expected:    `{"policies":null,"remediation-performed":true,"total-checks":0,"total-breaches":0,"remediation-totals":{"successful":1},"check-count-by-type":null,"breach-count-by-type":null,"breach-count-by-severity":null,"results":[{"name":"test-check","severity":"","check-type":"","passes":null,"breaches":[{"breach-type":"","check-type":"","check-name":"","severity":"","value":"Fail b","remediation":{"Status":"success","Messages":["fixed 1"]}}],"warnings":null,"status":"Fail","remediation-status":"","detailed-message":""}]}` + "\n",
			expectError: false,
			expectNil:   false,
		},
		{
			name: "junit format with mixed results",
			file: &File{
				Path:   "testdata/results.xml",
				Format: "junit",
			},
			rl: &result.ResultList{
				Policies: map[string][]string{"test-check": {"a", "b"}},
				Results: []result.Result{
					{
						Name:   "a",
						Status: result.Pass,
					},
					{
						Name:   "b",
						Status: result.Fail,
						Breaches: []breach.Breach{
							&breach.ValueBreach{Value: "Fail b"},
						},
						DetailedMessage: "Custom message for b",
					},
				},
			},
			expected: `<?xml version="1.0" encoding="UTF-8"?>
<testsuites tests="0" errors="0">
    <testsuite name="test-check" tests="0" errors="0">
        <testcase name="a" classname="test-check"></testcase>
        <testcase name="b" classname="test-check">
            <error message="Fail b">Custom message for b</error>
        </testcase>
    </testsuite>
</testsuites>
`,
			expectError: false,
			expectNil:   false,
		},
		{
			name: "unsupported format",
			file: &File{
				Path:   "testdata/unsupported.txt",
				Format: "unsupported",
			},
			rl: &result.ResultList{
				Results: []result.Result{
					{
						Name:   "test-check",
						Status: result.Pass,
					},
				},
			},
			expected:    "unsupported output format: unsupported",
			expectError: true,
			expectNil:   false,
		},
		{
			name: "empty path",
			file: &File{
				Path:   "",
				Format: "junit",
			},
			rl: &result.ResultList{
				Results: []result.Result{
					{
						Name:   "test-check",
						Status: result.Pass,
					},
				},
			},
			expected:    "",
			expectError: false,
			expectNil:   true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Create testdata directory if it doesn't exist
			if err := os.MkdirAll("testdata", 0755); err != nil {
				t.Fatal(err)
			}

			// Clean up test files after test
			defer func() {
				if err := os.RemoveAll("testdata"); err != nil {
					t.Logf("failed to clean up testdata directory: %v", err)
				}
			}()

			// Test the Output method
			output, err := tc.file.Output(tc.rl)

			if tc.expectError {
				assert.Error(err)
				assert.Contains(err.Error(), tc.expected)
				return
			}

			if tc.expectNil {
				assert.NoError(err)
				assert.Nil(output, "Output should be nil when path is empty")
				return
			}

			assert.NoError(err)
			assert.Equal(tc.expected, string(output))

			// Verify file was created
			_, err = os.Stat(tc.file.Path)
			assert.NoError(err)

			// Read file contents and verify
			fileContents, err := os.ReadFile(tc.file.Path)
			assert.NoError(err)
			assert.Equal(tc.expected, string(fileContents))
		})
	}
}

func TestFileOutputDirectoryCreation(t *testing.T) {
	tt := []struct {
		name     string
		file     *File
		expected string
	}{
		{
			name: "nested directory",
			file: &File{
				Path:   "testdata/nested/dir/results.xml",
				Format: "junit",
			},
			expected: "testdata/nested/dir",
		},
		{
			name: "current directory",
			file: &File{
				Path:   "results.xml",
				Format: "junit",
			},
			expected: ".",
		},
		{
			name: "deep nested directory",
			file: &File{
				Path:   "testdata/a/b/c/d/e/results.xml",
				Format: "junit",
			},
			expected: "testdata/a/b/c/d/e",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Clean up test files after test
			defer func() {
				if err := os.RemoveAll("testdata"); err != nil {
					t.Logf("failed to clean up testdata directory: %v", err)
				}
			}()

			// Test the Output method
			_, err := tc.file.Output(&result.ResultList{})
			assert.NoError(err)

			// Verify directory was created
			dir := filepath.Dir(tc.file.Path)
			_, err = os.Stat(dir)
			assert.NoError(err)
			assert.Equal(tc.expected, dir)
		})
	}
}

func TestFileOutputErrorHandling(t *testing.T) {
	tt := []struct {
		name        string
		file        *File
		rl          *result.ResultList
		expected    string
		expectError bool
	}{
		{
			name: "read-only directory",
			file: &File{
				Path:   "/readonly/results.xml",
				Format: "junit",
			},
			rl:          &result.ResultList{},
			expected:    "failed to create output directory",
			expectError: true,
		},
		{
			name: "invalid path",
			file: &File{
				Path:   "/dev/null/results.xml", // This path should be invalid on most systems
				Format: "junit",
			},
			rl:          &result.ResultList{},
			expected:    "failed to create output directory",
			expectError: true,
		},
		{
			name: "empty path",
			file: &File{
				Path:   "",
				Format: "junit",
			},
			rl:          &result.ResultList{},
			expected:    "", // No error expected, should return nil
			expectError: false,
		},
		{
			name: "invalid format",
			file: &File{
				Path:   "testdata/results.txt",
				Format: "invalid",
			},
			rl:          &result.ResultList{},
			expected:    "unsupported output format: invalid",
			expectError: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Create a read-only directory for testing
			if tc.name == "read-only directory" {
				if err := os.MkdirAll("/readonly", 0444); err != nil {
					t.Skip("requires root privileges to create read-only directory")
				}
				defer os.RemoveAll("/readonly")
			}

			// Test the Output method
			output, err := tc.file.Output(tc.rl)

			if tc.expectError {
				assert.Error(err)
				assert.Contains(err.Error(), tc.expected)
			} else {
				assert.NoError(err)
				assert.Nil(output, "Output should be nil when path is empty")
			}
		})
	}
}

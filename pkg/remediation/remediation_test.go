package remediation_test

import (
	"io"
	"math"
	"testing"

	"github.com/sirupsen/logrus"
	logrus_test "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"

	. "github.com/salsadigitalauorg/shipshape/pkg/remediation"
	"github.com/salsadigitalauorg/shipshape/pkg/remediation/testdata"
)

func TestRemediator(t *testing.T) {
	assert := assert.New(t)

	testrem := &testdata.TestRemediator{
		Message: "remediation message",
		ExpectedRemediationResult: RemediationResult{
			Status:   RemediationStatusFailed,
			Messages: []string{"foo"},
		},
	}
	assert.Implements((*Remediator)(nil), testrem)
	assert.Equal(testrem.PluginName(), "test")
	assert.Equal(testrem.GetRemediationMessage(), "remediation message")
	assert.Equal(RemediationResult{
		Status:   RemediationStatusFailed,
		Messages: []string{"foo"},
	}, testrem.Remediate())
}

func TestRemediatorFromInterface(t *testing.T) {
	assert := assert.New(t)

	tt := []struct {
		name                      string
		input                     interface{}
		expected                  Remediator
		expectFatal               string
		expectFatalHookEntryIndex int
		expectPanic               bool
	}{
		{
			name:     "nil",
			input:    nil,
			expected: nil,
		},
		{
			name:     "default",
			input:    map[string]any{},
			expected: &CommandRemediator{},
		},
		{
			name: "test",
			input: map[string]any{
				"plugin": "test",
				"expected-remediation-result": map[string]any{
					"status":   "failed",
					"messages": []string{"foo"},
				},
			},
			expected: &testdata.TestRemediator{
				ExpectedRemediationResult: RemediationResult{
					Status:   RemediationStatusFailed,
					Messages: []string{"foo"},
				},
			},
		},
		{
			name: "unknown",
			input: map[string]any{
				"plugin": "unknown",
			},
			expected:    nil,
			expectFatal: "unknown remediation plugin",
			expectPanic: true,
		},
		{
			name: "invalid/jsonMarshalError",
			input: map[string]any{
				"plugin": math.Inf(1),
			},
			expected:    nil,
			expectFatal: "json: unsupported value: +Inf",
		},
		{
			name: "invalid/jsonUnmarshalError/firstPass",
			input: map[string]any{
				"plugin": 999,
			},
			expected:    nil,
			expectFatal: "json: cannot unmarshal number into Go struct field .plugin of type string",
		},
		{
			name: "invalid/jsonUnmarshalError/final",
			input: map[string]any{
				"plugin":                      "test",
				"expected-remediation-result": "foo",
			},
			expected:    nil,
			expectFatal: "json: cannot unmarshal string into Go struct field TestRemediator.expected-remediation-result of type breach.RemediationResult",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			origRegistry := Registry
			defer func() { Registry = origRegistry }()
			Registry["test"] = func() Remediator { return &testdata.TestRemediator{} }

			// Hide logging output.
			currLogOut := logrus.StandardLogger().Out
			defer logrus.SetOutput(currLogOut)
			logrus.SetOutput(io.Discard)

			// Interrupt fatal exit.
			currExitFunc := logrus.StandardLogger().ExitFunc
			defer func() { logrus.StandardLogger().ExitFunc = currExitFunc }()
			logrus.StandardLogger().ExitFunc = func(int) {}

			// Install a test hook so we can test fatals.
			hook := logrus_test.Hook{}
			currHooks := logrus.StandardLogger().Hooks
			defer func() { logrus.StandardLogger().Hooks = currHooks }()
			logrus.StandardLogger().AddHook(&hook)

			if tc.expectFatal != "" {
				if tc.expectPanic {
					assert.Panics(func() { RemediatorFromInterface(tc.input) })
				} else {
					RemediatorFromInterface(tc.input)
				}
				assert.Equal(logrus.FatalLevel, hook.Entries[tc.expectFatalHookEntryIndex].Level)
				assert.Equal(tc.expectFatal, hook.Entries[tc.expectFatalHookEntryIndex].Message)
				return
			} else {
				assert.Equal(tc.expected, RemediatorFromInterface(tc.input))
			}
		})
	}
}

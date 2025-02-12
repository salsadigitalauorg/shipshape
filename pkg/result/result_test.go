package result_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/remediation"
	. "github.com/salsadigitalauorg/shipshape/pkg/result"
)

func TestResultSort(t *testing.T) {
	assert := assert.New(t)

	r := Result{
		Passes: []string{"z pass", "g pass", "a pass", "b pass"},
		Breaches: []breach.Breach{
			&breach.ValueBreach{CheckName: "x", Value: "breach 1"},
			&breach.ValueBreach{CheckName: "h", Value: "breach 2"},
			&breach.ValueBreach{CheckName: "v", Value: "breach 3"},
			&breach.ValueBreach{CheckName: "f", Value: "breach 4"},
		},
		Warnings: []string{"y warn", "i warn", "u warn", "c warn"},
	}
	r.Sort()

	assert.EqualValues(Result{
		Passes: []string{"a pass", "b pass", "g pass", "z pass"},
		Breaches: []breach.Breach{
			&breach.ValueBreach{CheckName: "f", Value: "breach 4"},
			&breach.ValueBreach{CheckName: "h", Value: "breach 2"},
			&breach.ValueBreach{CheckName: "v", Value: "breach 3"},
			&breach.ValueBreach{CheckName: "x", Value: "breach 1"},
		},
		Warnings: []string{"c warn", "i warn", "u warn", "y warn"},
	}, r)
}

func TestResultRemediationsCount(t *testing.T) {
	assert := assert.New(t)

	r := Result{
		Breaches: []breach.Breach{
			&breach.ValueBreach{CheckName: "x", RemediationResult: remediation.RemediationResult{Status: remediation.RemediationStatusNoSupport}},
			&breach.ValueBreach{CheckName: "h", RemediationResult: remediation.RemediationResult{Status: remediation.RemediationStatusSuccess}},
			&breach.ValueBreach{CheckName: "i", RemediationResult: remediation.RemediationResult{Status: remediation.RemediationStatusSuccess}},
			&breach.ValueBreach{CheckName: "v", RemediationResult: remediation.RemediationResult{Status: remediation.RemediationStatusFailed}},
			&breach.ValueBreach{CheckName: "w", RemediationResult: remediation.RemediationResult{Status: remediation.RemediationStatusFailed}},
			&breach.ValueBreach{CheckName: "x", RemediationResult: remediation.RemediationResult{Status: remediation.RemediationStatusFailed}},
			&breach.ValueBreach{CheckName: "f", RemediationResult: remediation.RemediationResult{Status: remediation.RemediationStatusPartial}},
			&breach.ValueBreach{CheckName: "e", RemediationResult: remediation.RemediationResult{Status: remediation.RemediationStatusPartial}},
			&breach.ValueBreach{CheckName: "d", RemediationResult: remediation.RemediationResult{Status: remediation.RemediationStatusPartial}},
			&breach.ValueBreach{CheckName: "c", RemediationResult: remediation.RemediationResult{Status: remediation.RemediationStatusPartial}},
		},
	}
	unsupported, successful, failed, partial := r.RemediationsCount()
	assert.EqualValues(1, unsupported)
	assert.EqualValues(2, successful)
	assert.EqualValues(3, failed)
	assert.EqualValues(4, partial)
}

func TestResultDetermineResultStatus(t *testing.T) {
	tt := []struct {
		name                      string
		remediationPerformed      bool
		breaches                  []breach.Breach
		expectedStatus            Status
		expectedRemediationStatus remediation.RemediationStatus
	}{
		{
			name:                      "noBreach",
			remediationPerformed:      false,
			breaches:                  []breach.Breach{},
			expectedStatus:            Pass,
			expectedRemediationStatus: "",
		},
		{
			name:                      "noBreachRemediation",
			remediationPerformed:      true,
			breaches:                  []breach.Breach{},
			expectedStatus:            Pass,
			expectedRemediationStatus: remediation.RemediationStatusSuccess,
		},

		// Single breach.
		{
			name:                 "singleBreach",
			remediationPerformed: false,
			breaches: []breach.Breach{
				&breach.ValueBreach{CheckName: "x", Value: "breach 1"},
			},
			expectedStatus:            Fail,
			expectedRemediationStatus: "",
		},
		{
			name:                 "singleBreachRemediationNotSupported",
			remediationPerformed: true,
			breaches: []breach.Breach{
				&breach.ValueBreach{
					CheckName:         "x",
					Value:             "breach 1",
					RemediationResult: remediation.RemediationResult{Status: remediation.RemediationStatusNoSupport},
				},
			},
			expectedStatus:            Fail,
			expectedRemediationStatus: remediation.RemediationStatusNoSupport,
		},
		{
			name:                 "singleBreachRemediationSuccess",
			remediationPerformed: true,
			breaches: []breach.Breach{
				&breach.ValueBreach{
					CheckName:         "x",
					Value:             "breach 1",
					RemediationResult: remediation.RemediationResult{Status: remediation.RemediationStatusSuccess},
				},
			},
			expectedStatus:            Pass,
			expectedRemediationStatus: remediation.RemediationStatusSuccess,
		},
		{
			name:                 "singleBreachRemediationFailed",
			remediationPerformed: true,
			breaches: []breach.Breach{
				&breach.ValueBreach{
					CheckName:         "x",
					Value:             "breach 1",
					RemediationResult: remediation.RemediationResult{Status: remediation.RemediationStatusFailed},
				},
			},
			expectedStatus:            Fail,
			expectedRemediationStatus: remediation.RemediationStatusFailed,
		},
		{
			name:                 "singleBreachRemediationPartial",
			remediationPerformed: true,
			breaches: []breach.Breach{
				&breach.ValueBreach{
					CheckName:         "x",
					Value:             "breach 1",
					RemediationResult: remediation.RemediationResult{Status: remediation.RemediationStatusPartial},
				},
			},
			expectedStatus:            Fail,
			expectedRemediationStatus: remediation.RemediationStatusPartial,
		},

		// Multiple breaches.
		{
			name:                 "multipleBreaches",
			remediationPerformed: false,
			breaches: []breach.Breach{
				&breach.ValueBreach{CheckName: "x", Value: "breach 1"},
				&breach.ValueBreach{CheckName: "f", Value: "breach 2"},
			},
			expectedStatus:            Fail,
			expectedRemediationStatus: "",
		},
		{
			name:                 "multipleBreachesRemediationFailed",
			remediationPerformed: true,
			breaches: []breach.Breach{
				&breach.ValueBreach{
					CheckName:         "x",
					Value:             "breach 1",
					RemediationResult: remediation.RemediationResult{Status: remediation.RemediationStatusFailed},
				},
				&breach.ValueBreach{
					CheckName:         "f",
					Value:             "breach 2",
					RemediationResult: remediation.RemediationResult{Status: remediation.RemediationStatusFailed},
				},
			},
			expectedStatus:            Fail,
			expectedRemediationStatus: remediation.RemediationStatusFailed,
		},
		{
			name:                 "multipleBreachesRemediationUnsupported",
			remediationPerformed: true,
			breaches: []breach.Breach{
				&breach.ValueBreach{
					CheckName:         "x",
					Value:             "breach 1",
					RemediationResult: remediation.RemediationResult{Status: remediation.RemediationStatusNoSupport},
				},
				&breach.ValueBreach{
					CheckName:         "f",
					Value:             "breach 2",
					RemediationResult: remediation.RemediationResult{Status: remediation.RemediationStatusNoSupport},
				},
			},
			expectedStatus:            Fail,
			expectedRemediationStatus: remediation.RemediationStatusNoSupport,
		},

		// Multiple breaches with partial breach.Remediation.
		{
			name:                 "multipleBreachesRemediationPartial",
			remediationPerformed: true,
			breaches: []breach.Breach{
				&breach.ValueBreach{
					CheckName:         "x",
					Value:             "breach 1",
					RemediationResult: remediation.RemediationResult{Status: remediation.RemediationStatusSuccess},
				},
				&breach.ValueBreach{
					CheckName:         "f",
					Value:             "breach 2",
					RemediationResult: remediation.RemediationResult{Status: remediation.RemediationStatusFailed},
				},
			},
			expectedStatus:            Fail,
			expectedRemediationStatus: remediation.RemediationStatusPartial,
		},
		{
			name:                 "multipleBreachesRemediationPartial",
			remediationPerformed: true,
			breaches: []breach.Breach{
				&breach.ValueBreach{
					CheckName:         "x",
					Value:             "breach 1",
					RemediationResult: remediation.RemediationResult{Status: remediation.RemediationStatusSuccess},
				},
				&breach.ValueBreach{
					CheckName:         "f",
					Value:             "breach 2",
					RemediationResult: remediation.RemediationResult{Status: remediation.RemediationStatusFailed},
				},
			},
			expectedStatus:            Fail,
			expectedRemediationStatus: remediation.RemediationStatusPartial,
		},
		{
			name:                 "multipleBreachesRemediationPartial",
			remediationPerformed: true,
			breaches: []breach.Breach{
				&breach.ValueBreach{
					CheckName:         "x",
					Value:             "breach 1",
					RemediationResult: remediation.RemediationResult{Status: remediation.RemediationStatusPartial},
				},
				&breach.ValueBreach{
					CheckName:         "f",
					Value:             "breach 2",
					RemediationResult: remediation.RemediationResult{Status: remediation.RemediationStatusPartial},
				},
			},
			expectedStatus:            Fail,
			expectedRemediationStatus: remediation.RemediationStatusPartial,
		},
		{
			name:                 "multipleBreachesRemediationPartial",
			remediationPerformed: true,
			breaches: []breach.Breach{
				&breach.ValueBreach{
					CheckName:         "x",
					Value:             "breach 1",
					RemediationResult: remediation.RemediationResult{Status: remediation.RemediationStatusSuccess},
				},
				&breach.ValueBreach{
					CheckName:         "f",
					Value:             "breach 2",
					RemediationResult: remediation.RemediationResult{Status: remediation.RemediationStatusNoSupport},
				},
			},
			expectedStatus:            Fail,
			expectedRemediationStatus: remediation.RemediationStatusPartial,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			r := Result{Breaches: tc.breaches}
			r.DetermineResultStatus(tc.remediationPerformed)

			assert.Equal(tc.expectedStatus, r.Status)
			assert.Equal(tc.expectedRemediationStatus, r.RemediationStatus)
		})
	}
}

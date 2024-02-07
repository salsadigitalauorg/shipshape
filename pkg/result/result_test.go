package result_test

import (
	"testing"

	. "github.com/salsadigitalauorg/shipshape/pkg/result"

	"github.com/stretchr/testify/assert"
)

func TestResultSort(t *testing.T) {
	assert := assert.New(t)

	r := Result{
		Passes: []string{"z pass", "g pass", "a pass", "b pass"},
		Breaches: []Breach{
			&ValueBreach{CheckName: "x", Value: "breach 1"},
			&ValueBreach{CheckName: "h", Value: "breach 2"},
			&ValueBreach{CheckName: "v", Value: "breach 3"},
			&ValueBreach{CheckName: "f", Value: "breach 4"},
		},
		Warnings: []string{"y warn", "i warn", "u warn", "c warn"},
	}
	r.Sort()

	assert.EqualValues(Result{
		Passes: []string{"a pass", "b pass", "g pass", "z pass"},
		Breaches: []Breach{
			&ValueBreach{CheckName: "f", Value: "breach 4"},
			&ValueBreach{CheckName: "h", Value: "breach 2"},
			&ValueBreach{CheckName: "v", Value: "breach 3"},
			&ValueBreach{CheckName: "x", Value: "breach 1"},
		},
		Warnings: []string{"c warn", "i warn", "u warn", "y warn"},
	}, r)
}

func TestResultRemediationsCount(t *testing.T) {
	assert := assert.New(t)

	r := Result{
		Breaches: []Breach{
			&ValueBreach{CheckName: "x", Remediation: Remediation{Status: RemediationStatusNoSupport}},
			&ValueBreach{CheckName: "h", Remediation: Remediation{Status: RemediationStatusSuccess}},
			&ValueBreach{CheckName: "i", Remediation: Remediation{Status: RemediationStatusSuccess}},
			&ValueBreach{CheckName: "v", Remediation: Remediation{Status: RemediationStatusFailed}},
			&ValueBreach{CheckName: "w", Remediation: Remediation{Status: RemediationStatusFailed}},
			&ValueBreach{CheckName: "x", Remediation: Remediation{Status: RemediationStatusFailed}},
			&ValueBreach{CheckName: "f", Remediation: Remediation{Status: RemediationStatusPartial}},
			&ValueBreach{CheckName: "e", Remediation: Remediation{Status: RemediationStatusPartial}},
			&ValueBreach{CheckName: "d", Remediation: Remediation{Status: RemediationStatusPartial}},
			&ValueBreach{CheckName: "c", Remediation: Remediation{Status: RemediationStatusPartial}},
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
		breaches                  []Breach
		expectedStatus            Status
		expectedRemediationStatus RemediationStatus
	}{
		{
			name:                      "noBreach",
			remediationPerformed:      false,
			breaches:                  []Breach{},
			expectedStatus:            Pass,
			expectedRemediationStatus: "",
		},
		{
			name:                      "noBreachRemediation",
			remediationPerformed:      true,
			breaches:                  []Breach{},
			expectedStatus:            Pass,
			expectedRemediationStatus: RemediationStatusSuccess,
		},

		// Single breach.
		{
			name:                 "singleBreach",
			remediationPerformed: false,
			breaches: []Breach{
				&ValueBreach{CheckName: "x", Value: "breach 1"},
			},
			expectedStatus:            Fail,
			expectedRemediationStatus: "",
		},
		{
			name:                 "singleBreachRemediationNotSupported",
			remediationPerformed: true,
			breaches: []Breach{
				&ValueBreach{
					CheckName:   "x",
					Value:       "breach 1",
					Remediation: Remediation{Status: RemediationStatusNoSupport},
				},
			},
			expectedStatus:            Fail,
			expectedRemediationStatus: RemediationStatusNoSupport,
		},
		{
			name:                 "singleBreachRemediationSuccess",
			remediationPerformed: true,
			breaches: []Breach{
				&ValueBreach{
					CheckName:   "x",
					Value:       "breach 1",
					Remediation: Remediation{Status: RemediationStatusSuccess},
				},
			},
			expectedStatus:            Pass,
			expectedRemediationStatus: RemediationStatusSuccess,
		},
		{
			name:                 "singleBreachRemediationFailed",
			remediationPerformed: true,
			breaches: []Breach{
				&ValueBreach{
					CheckName:   "x",
					Value:       "breach 1",
					Remediation: Remediation{Status: RemediationStatusFailed},
				},
			},
			expectedStatus:            Fail,
			expectedRemediationStatus: RemediationStatusFailed,
		},
		{
			name:                 "singleBreachRemediationPartial",
			remediationPerformed: true,
			breaches: []Breach{
				&ValueBreach{
					CheckName:   "x",
					Value:       "breach 1",
					Remediation: Remediation{Status: RemediationStatusPartial},
				},
			},
			expectedStatus:            Fail,
			expectedRemediationStatus: RemediationStatusPartial,
		},

		// Multiple breaches.
		{
			name:                 "multipleBreaches",
			remediationPerformed: false,
			breaches: []Breach{
				&ValueBreach{CheckName: "x", Value: "breach 1"},
				&ValueBreach{CheckName: "f", Value: "breach 2"},
			},
			expectedStatus:            Fail,
			expectedRemediationStatus: "",
		},
		{
			name:                 "multipleBreachesRemediationFailed",
			remediationPerformed: true,
			breaches: []Breach{
				&ValueBreach{
					CheckName:   "x",
					Value:       "breach 1",
					Remediation: Remediation{Status: RemediationStatusFailed},
				},
				&ValueBreach{
					CheckName:   "f",
					Value:       "breach 2",
					Remediation: Remediation{Status: RemediationStatusFailed},
				},
			},
			expectedStatus:            Fail,
			expectedRemediationStatus: RemediationStatusFailed,
		},
		{
			name:                 "multipleBreachesRemediationUnsupported",
			remediationPerformed: true,
			breaches: []Breach{
				&ValueBreach{
					CheckName:   "x",
					Value:       "breach 1",
					Remediation: Remediation{Status: RemediationStatusNoSupport},
				},
				&ValueBreach{
					CheckName:   "f",
					Value:       "breach 2",
					Remediation: Remediation{Status: RemediationStatusNoSupport},
				},
			},
			expectedStatus:            Fail,
			expectedRemediationStatus: RemediationStatusNoSupport,
		},

		// Multiple breaches with partial remediation.
		{
			name:                 "multipleBreachesRemediationPartial",
			remediationPerformed: true,
			breaches: []Breach{
				&ValueBreach{
					CheckName:   "x",
					Value:       "breach 1",
					Remediation: Remediation{Status: RemediationStatusSuccess},
				},
				&ValueBreach{
					CheckName:   "f",
					Value:       "breach 2",
					Remediation: Remediation{Status: RemediationStatusFailed},
				},
			},
			expectedStatus:            Fail,
			expectedRemediationStatus: RemediationStatusPartial,
		},
		{
			name:                 "multipleBreachesRemediationPartial",
			remediationPerformed: true,
			breaches: []Breach{
				&ValueBreach{
					CheckName:   "x",
					Value:       "breach 1",
					Remediation: Remediation{Status: RemediationStatusSuccess},
				},
				&ValueBreach{
					CheckName:   "f",
					Value:       "breach 2",
					Remediation: Remediation{Status: RemediationStatusFailed},
				},
			},
			expectedStatus:            Fail,
			expectedRemediationStatus: RemediationStatusPartial,
		},
		{
			name:                 "multipleBreachesRemediationPartial",
			remediationPerformed: true,
			breaches: []Breach{
				&ValueBreach{
					CheckName:   "x",
					Value:       "breach 1",
					Remediation: Remediation{Status: RemediationStatusPartial},
				},
				&ValueBreach{
					CheckName:   "f",
					Value:       "breach 2",
					Remediation: Remediation{Status: RemediationStatusPartial},
				},
			},
			expectedStatus:            Fail,
			expectedRemediationStatus: RemediationStatusPartial,
		},
		{
			name:                 "multipleBreachesRemediationPartial",
			remediationPerformed: true,
			breaches: []Breach{
				&ValueBreach{
					CheckName:   "x",
					Value:       "breach 1",
					Remediation: Remediation{Status: RemediationStatusSuccess},
				},
				&ValueBreach{
					CheckName:   "f",
					Value:       "breach 2",
					Remediation: Remediation{Status: RemediationStatusNoSupport},
				},
			},
			expectedStatus:            Fail,
			expectedRemediationStatus: RemediationStatusPartial,
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

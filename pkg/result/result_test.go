package result_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
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
			&breach.ValueBreach{CheckName: "x", RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusNoSupport}},
			&breach.ValueBreach{CheckName: "h", RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusSuccess}},
			&breach.ValueBreach{CheckName: "i", RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusSuccess}},
			&breach.ValueBreach{CheckName: "v", RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusFailed}},
			&breach.ValueBreach{CheckName: "w", RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusFailed}},
			&breach.ValueBreach{CheckName: "x", RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusFailed}},
			&breach.ValueBreach{CheckName: "f", RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusPartial}},
			&breach.ValueBreach{CheckName: "e", RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusPartial}},
			&breach.ValueBreach{CheckName: "d", RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusPartial}},
			&breach.ValueBreach{CheckName: "c", RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusPartial}},
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
		expectedRemediationStatus breach.RemediationStatus
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
			expectedRemediationStatus: breach.RemediationStatusSuccess,
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
					RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusNoSupport},
				},
			},
			expectedStatus:            Fail,
			expectedRemediationStatus: breach.RemediationStatusNoSupport,
		},
		{
			name:                 "singleBreachRemediationSuccess",
			remediationPerformed: true,
			breaches: []breach.Breach{
				&breach.ValueBreach{
					CheckName:         "x",
					Value:             "breach 1",
					RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusSuccess},
				},
			},
			expectedStatus:            Pass,
			expectedRemediationStatus: breach.RemediationStatusSuccess,
		},
		{
			name:                 "singleBreachRemediationFailed",
			remediationPerformed: true,
			breaches: []breach.Breach{
				&breach.ValueBreach{
					CheckName:         "x",
					Value:             "breach 1",
					RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusFailed},
				},
			},
			expectedStatus:            Fail,
			expectedRemediationStatus: breach.RemediationStatusFailed,
		},
		{
			name:                 "singleBreachRemediationPartial",
			remediationPerformed: true,
			breaches: []breach.Breach{
				&breach.ValueBreach{
					CheckName:         "x",
					Value:             "breach 1",
					RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusPartial},
				},
			},
			expectedStatus:            Fail,
			expectedRemediationStatus: breach.RemediationStatusPartial,
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
					RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusFailed},
				},
				&breach.ValueBreach{
					CheckName:         "f",
					Value:             "breach 2",
					RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusFailed},
				},
			},
			expectedStatus:            Fail,
			expectedRemediationStatus: breach.RemediationStatusFailed,
		},
		{
			name:                 "multipleBreachesRemediationUnsupported",
			remediationPerformed: true,
			breaches: []breach.Breach{
				&breach.ValueBreach{
					CheckName:         "x",
					Value:             "breach 1",
					RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusNoSupport},
				},
				&breach.ValueBreach{
					CheckName:         "f",
					Value:             "breach 2",
					RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusNoSupport},
				},
			},
			expectedStatus:            Fail,
			expectedRemediationStatus: breach.RemediationStatusNoSupport,
		},

		// Multiple breaches with partial breach.Remediation.
		{
			name:                 "multipleBreachesRemediationPartial",
			remediationPerformed: true,
			breaches: []breach.Breach{
				&breach.ValueBreach{
					CheckName:         "x",
					Value:             "breach 1",
					RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusSuccess},
				},
				&breach.ValueBreach{
					CheckName:         "f",
					Value:             "breach 2",
					RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusFailed},
				},
			},
			expectedStatus:            Fail,
			expectedRemediationStatus: breach.RemediationStatusPartial,
		},
		{
			name:                 "multipleBreachesRemediationPartial",
			remediationPerformed: true,
			breaches: []breach.Breach{
				&breach.ValueBreach{
					CheckName:         "x",
					Value:             "breach 1",
					RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusSuccess},
				},
				&breach.ValueBreach{
					CheckName:         "f",
					Value:             "breach 2",
					RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusFailed},
				},
			},
			expectedStatus:            Fail,
			expectedRemediationStatus: breach.RemediationStatusPartial,
		},
		{
			name:                 "multipleBreachesRemediationPartial",
			remediationPerformed: true,
			breaches: []breach.Breach{
				&breach.ValueBreach{
					CheckName:         "x",
					Value:             "breach 1",
					RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusPartial},
				},
				&breach.ValueBreach{
					CheckName:         "f",
					Value:             "breach 2",
					RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusPartial},
				},
			},
			expectedStatus:            Fail,
			expectedRemediationStatus: breach.RemediationStatusPartial,
		},
		{
			name:                 "multipleBreachesRemediationPartial",
			remediationPerformed: true,
			breaches: []breach.Breach{
				&breach.ValueBreach{
					CheckName:         "x",
					Value:             "breach 1",
					RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusSuccess},
				},
				&breach.ValueBreach{
					CheckName:         "f",
					Value:             "breach 2",
					RemediationResult: breach.RemediationResult{Status: breach.RemediationStatusNoSupport},
				},
			},
			expectedStatus:            Fail,
			expectedRemediationStatus: breach.RemediationStatusPartial,
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

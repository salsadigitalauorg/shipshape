package result_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/salsadigitalauorg/shipshape/pkg/result"
)

func TestBreachSetCommonValues(t *testing.T) {
	assert := assert.New(t)

	type bogusBreach struct{}

	tests := []struct {
		name      string
		breach    Breach
		checkType string
		checkName string
		severity  string
		empty     bool
	}{
		{
			name:      "ValueBreach",
			breach:    ValueBreach{},
			checkType: "ctvb",
			checkName: "valuebreachcheck",
			severity:  "low",
		},
		{
			name:      "KeyValueBreach",
			breach:    KeyValueBreach{},
			checkType: "ctkvb",
			checkName: "keyvaluebreachcheck",
			severity:  "normal",
		},
		{
			name:      "KeyValuesBreach",
			breach:    KeyValuesBreach{},
			checkType: "ctkvsb",
			checkName: "keyvaluesbreachcheck",
			severity:  "high",
		},
		{
			name:      "BogusBreach",
			breach:    bogusBreach{},
			checkType: "ctbb",
			checkName: "bogusbreachcheck",
			severity:  "critical",
			empty:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			BreachSetCommonValues(&test.breach, test.checkType, test.checkName,
				test.severity)
			if !test.empty {
				assert.Equal(test.checkName, BreachGetCheckName(test.breach))
				assert.Equal(test.checkType, BreachGetCheckType(test.breach))
				assert.Equal(test.severity, BreachGetSeverity(test.breach))
			} else {
				assert.Equal("", BreachGetCheckName(test.breach))
				assert.Equal("", BreachGetCheckType(test.breach))
				assert.Equal("", BreachGetSeverity(test.breach))
			}
		})
	}
}

func TestBreachGetters(t *testing.T) {
	assert := assert.New(t)

	type bogusBreach struct{}

	tests := []struct {
		name               string
		breach             Breach
		expectedKeyLabel   string
		expectedKey        string
		expectedValueLabel string
		expectedValue      string
		expectedValues     []string
	}{
		{
			name: "ValueBreach",
			breach: ValueBreach{
				ValueLabel: "vbvl",
				Value:      "vbv",
			},
			expectedKeyLabel:   "",
			expectedKey:        "",
			expectedValueLabel: "vbvl",
			expectedValue:      "vbv",
			expectedValues:     []string(nil),
		},
		{
			name: "KeyValueBreach",
			breach: KeyValueBreach{
				KeyLabel:   "kvbklbl",
				Key:        "kvbk",
				ValueLabel: "kvbvl",
				Value:      "kvbv",
			},
			expectedKeyLabel:   "kvbklbl",
			expectedKey:        "kvbk",
			expectedValueLabel: "kvbvl",
			expectedValue:      "kvbv",
			expectedValues:     []string(nil),
		},
		{
			name: "KeyValuesBreach",
			breach: KeyValuesBreach{
				KeyLabel:   "kvsbklbl",
				Key:        "kvsbk",
				ValueLabel: "kvsbvl",
				Values:     []string{"kvsbvs1"},
			},
			expectedKeyLabel:   "kvsbklbl",
			expectedKey:        "kvsbk",
			expectedValueLabel: "kvsbvl",
			expectedValue:      "",
			expectedValues:     []string{"kvsbvs1"},
		},
		{
			name:               "bogusBreach",
			breach:             bogusBreach{},
			expectedKeyLabel:   "",
			expectedKey:        "",
			expectedValueLabel: "",
			expectedValue:      "",
			expectedValues:     []string(nil),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(test.expectedKeyLabel, BreachGetKeyLabel(test.breach))
			assert.Equal(test.expectedKey, BreachGetKey(test.breach))
			assert.Equal(test.expectedValueLabel, BreachGetValueLabel(test.breach))
			assert.Equal(test.expectedValue, BreachGetValue(test.breach))
			assert.EqualValues(test.expectedValues, BreachGetValues(test.breach))
		})
	}
}

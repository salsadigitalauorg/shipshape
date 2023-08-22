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
		name               string
		breach             Breach
		expectedBreachType BreachType
		expectedCheckType  string
		expectedCheckName  string
		expectedSeverity   string
		empty              bool
	}{
		{
			name:               "ValueBreach",
			breach:             ValueBreach{},
			expectedBreachType: BreachTypeValue,
			expectedCheckType:  "ctvb",
			expectedCheckName:  "valuebreachcheck",
			expectedSeverity:   "low",
		},
		{
			name:               "KeyValueBreach",
			breach:             KeyValueBreach{},
			expectedBreachType: BreachTypeKeyValue,
			expectedCheckType:  "ctkvb",
			expectedCheckName:  "keyvaluebreachcheck",
			expectedSeverity:   "normal",
		},
		{
			name:               "KeyValuesBreach",
			breach:             KeyValuesBreach{},
			expectedBreachType: BreachTypeKeyValues,
			expectedCheckType:  "ctkvsb",
			expectedCheckName:  "keyvaluesbreachcheck",
			expectedSeverity:   "high",
		},
		{
			name:               "BogusBreach",
			breach:             bogusBreach{},
			expectedBreachType: "",
			expectedCheckType:  "ctbb",
			expectedCheckName:  "bogusbreachcheck",
			expectedSeverity:   "critical",
			empty:              true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			BreachSetCommonValues(&test.breach, test.expectedCheckType, test.expectedCheckName,
				test.expectedSeverity)
			if !test.empty {
				assert.Equal(test.expectedBreachType, BreachGetBreachType(test.breach))
				assert.Equal(test.expectedCheckName, BreachGetCheckName(test.breach))
				assert.Equal(test.expectedCheckType, BreachGetCheckType(test.breach))
				assert.Equal(test.expectedSeverity, BreachGetSeverity(test.breach))
			} else {
				assert.Equal(BreachType(""), BreachGetBreachType(test.breach))
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
		name                  string
		breach                Breach
		expectedKeyLabel      string
		expectedKey           string
		expectedValueLabel    string
		expectedValue         string
		expectedValues        []string
		expectedExpectedValue string
	}{
		{
			name: "ValueBreach",
			breach: ValueBreach{
				ValueLabel:    "vbvl",
				Value:         "vbv",
				ExpectedValue: "vbve",
			},
			expectedKeyLabel:      "",
			expectedKey:           "",
			expectedValueLabel:    "vbvl",
			expectedValue:         "vbv",
			expectedValues:        []string(nil),
			expectedExpectedValue: "vbve",
		},
		{
			name: "KeyValueBreach",
			breach: KeyValueBreach{
				KeyLabel:      "kvbklbl",
				Key:           "kvbk",
				ValueLabel:    "kvbvl",
				Value:         "kvbv",
				ExpectedValue: "kvbve",
			},
			expectedKeyLabel:      "kvbklbl",
			expectedKey:           "kvbk",
			expectedValueLabel:    "kvbvl",
			expectedValue:         "kvbv",
			expectedValues:        []string(nil),
			expectedExpectedValue: "kvbve",
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
			assert.Equal(test.expectedExpectedValue, BreachGetExpectedValue(test.breach))
			assert.EqualValues(test.expectedValues, BreachGetValues(test.breach))
		})
	}
}

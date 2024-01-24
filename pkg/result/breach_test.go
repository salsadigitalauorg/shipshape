package result_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/salsadigitalauorg/shipshape/pkg/result"
)

func TestBreachValueBreachStringer(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		name     string
		breach   Breach
		expected string
	}{
		{
			name: "value-breach",
			breach: ValueBreach{
				ValueLabel: "file not found",
				Value:      "foo.ext",
			},
			expected: "[file not found] foo.ext",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(test.expected, test.breach.String())
		})
	}
}

func TestBreachKeyValueBreachStringer(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		name     string
		breach   Breach
		expected string
	}{
		{
			name: "key-value-breach-1",
			breach: KeyValueBreach{
				KeyLabel:   "config",
				Key:        "clamav.settings",
				ValueLabel: "key not found",
				Value:      "enabled",
			},
			expected: "[config clamav.settings] key not found: enabled",
		},
		{
			name: "key-value-breach-2",
			breach: KeyValueBreach{
				KeyLabel:      "clamav.settings",
				Key:           "enabled",
				Value:         "false",
				ExpectedValue: "true",
			},
			expected: "[clamav.settings] 'enabled' equals 'false', expected 'true'",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(test.expected, test.breach.String())
		})
	}
}

func TestBreachKeyValuesBreachStringers(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		name     string
		breach   Breach
		expected string
	}{
		{
			name: "KeyValuesBreach",
			breach: KeyValuesBreach{
				KeyLabel:   "role",
				Key:        "admin",
				ValueLabel: "disallowed permissions",
				Values:     []string{"delete the site", "delete the world"},
			},
			expected: "[role:admin] disallowed permissions: [delete the site, delete the world]",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(test.expected, test.breach.String())
		})
	}
}

type bogusBreach struct{}

func (b bogusBreach) String() string {
	return ""
}

func TestBreachSetCommonValues(t *testing.T) {
	assert := assert.New(t)

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

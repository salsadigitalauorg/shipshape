package drupal_test

import (
	"testing"

	. "github.com/salsadigitalauorg/shipshape/pkg/checks/drupal"
	"github.com/salsadigitalauorg/shipshape/pkg/checks/yaml"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
	"github.com/stretchr/testify/assert"
)

func TestTrackingCodeMerge(t *testing.T) {
	assert := assert.New(t)

	c := TrackingCodeCheck{
		DrushYamlCheck: DrushYamlCheck{
			YamlBase: yaml.YamlBase{
				Values: []yaml.KeyValue{
					{Key: "key1", Value: "val1", Optional: false},
				},
			},
		},
		Code:        "foo",
		DrushStatus: DrushStatus{Uri: "http://foo.example"},
	}
	c.Merge(&TrackingCodeCheck{
		DrushYamlCheck: DrushYamlCheck{
			YamlBase: yaml.YamlBase{
				Values: []yaml.KeyValue{
					{Key: "key1", Value: "val1", Optional: true},
				},
			},
		},
		Code:        "bar",
		DrushStatus: DrushStatus{Uri: "http://bar.example"},
	})
	assert.EqualValues(TrackingCodeCheck{
		DrushYamlCheck: DrushYamlCheck{
			YamlBase: yaml.YamlBase{
				Values: []yaml.KeyValue{
					{Key: "key1", Value: "val1", Optional: true},
				},
			},
		},
		Code:        "bar",
		DrushStatus: DrushStatus{Uri: "http://bar.example"},
	}, c)
}

func TestTrackingCodeUnmarshalData(t *testing.T) {
	assert := assert.New(t)

	c := TrackingCodeCheck{}
	c.ConfigName = "status"
	c.DataMap = map[string][]byte{
		"status": []byte(`
foo: bar

`),
	}
	c.UnmarshalDataMap()
	assert.NotEqual(result.Fail, c.Result.Status)
	assert.Equal("", c.DrushStatus.Uri)

	c.DataMap = map[string][]byte{
		"status": []byte(`
uri: https://foo.example

`),
	}
	c.UnmarshalDataMap()
	assert.NotEqual(result.Fail, c.Result.Status)
	assert.Equal("https://foo.example", c.DrushStatus.Uri)
}

func TestTrackingCodeCheckFails(t *testing.T) {
	assert := assert.New(t)

	c := TrackingCodeCheck{
		Code: "UA-xxxxxx-1",
	}
	c.Init(TrackingCode)
	assert.Equal("status", c.Command)

	c.DrushStatus = DrushStatus{
		Uri: "https://google.com",
	}
	c.RunCheck()
	c.Result.DetermineResultStatus(false)
	assert.Equal(result.Fail, c.Result.Status)
	assert.ElementsMatch(
		[]result.Breach{&result.KeyValueBreach{
			BreachType: "key-value",
			CheckType:  "drupal-tracking-code",
			Severity:   "normal",
			Key:        "required tracking code not present",
			Value:      "UA-xxxxxx-1"}},
		c.Result.Breaches,
	)
}

func TestTrackingCodeCheckPass(t *testing.T) {
	assert := assert.New(t)

	c := TrackingCodeCheck{
		Code: "UA-xxxxxx-1",
	}
	c.Init(TrackingCode)
	assert.Equal("status", c.Command)

	c.DrushStatus = DrushStatus{
		Uri: "https://gist.github.com/Pominova/cf7884e7418f6ebfa412d2d3dc472a97",
	}
	c.RunCheck()
	c.Result.DetermineResultStatus(false)
	assert.Equal(result.Pass, c.Result.Status)
	assert.ElementsMatch(
		[]string{"tracking code [UA-xxxxxx-1] present"},
		c.Result.Passes,
	)
}

package shipshape_test

import (
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/stretchr/testify/assert"
)

func TestYamlBaseMerge(t *testing.T) {
	assert := assert.New(t)

	c := shipshape.YamlBase{
		Values: []shipshape.KeyValue{{Key: "foo", Value: "bar"}},
	}
	err := c.Merge(&shipshape.YamlBase{
		Values: []shipshape.KeyValue{{Key: "baz", Value: "zoom"}},
	})
	assert.Equal(nil, err)
	assert.EqualValues(
		[]shipshape.KeyValue{{Key: "baz", Value: "zoom"}},
		c.Values,
	)
	err = c.Merge(&shipshape.YamlBase{
		Values: []shipshape.KeyValue{{Key: "baz", Value: "zap"}},
	})
	assert.Equal(nil, err)
	assert.EqualValues(
		[]shipshape.KeyValue{{Key: "baz", Value: "zap"}},
		c.Values,
	)
}

func TestYamlUnmarshalDataMap(t *testing.T) {
	assert := assert.New(t)

	// Invalid data.
	c := shipshape.YamlBase{
		CheckBase: shipshape.CheckBase{
			DataMap: map[string][]byte{
				"data": []byte(`
foo:
  bar:
	baz
`),
			},
		},
	}
	c.UnmarshalDataMap()
	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.EqualValues(0, len(c.Result.Passes))
	assert.EqualValues(
		[]string{"yaml: line 4: found character that cannot start any token"},
		c.Result.Failures)

	// Valid data.
	c = shipshape.YamlBase{
		CheckBase: shipshape.CheckBase{
			DataMap: map[string][]byte{
				"data": []byte(`
foo:
  bar:
    - name: baz
      value: zoom
`),
			},
		},
	}
	c.UnmarshalDataMap()
	assert.EqualValues(0, len(c.Result.Failures))

	// Invalid yaml kec.
	c = shipshape.YamlBase{
		CheckBase: shipshape.CheckBase{
			DataMap: map[string][]byte{
				"data": []byte(`
foo:
  bar:
    baz&*zoom: zap
`),
			},
		},
		Values: []shipshape.KeyValue{
			{Key: "baz&*zoom", Value: "zap"},
		},
	}
	c.RunCheck(false)

	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.EqualValues(
		[]string{"invalid character '&' at position 3, following \"baz\""},
		c.Result.Failures)
}

func TestYamlCheckKeyValue(t *testing.T) {
	assert := assert.New(t)

	c := shipshape.YamlBase{
		CheckBase: shipshape.CheckBase{
			DataMap: map[string][]byte{
				"data": []byte(`
foo:
  bar:
    - baz: zoo

`),
			},
		},
	}
	c.UnmarshalDataMap()

	// Invalid path.
	_, _, err := c.CheckKeyValue(shipshape.KeyValue{
		Key:   "&*&^);",
		Value: "foo",
	}, "data")
	assert.Equal("child name missing at position 0, following \"\"", err.Error())

	// Non-existent path.
	kvr, _, err := c.CheckKeyValue(shipshape.KeyValue{
		Key:   "foo.baz",
		Value: "foo",
	}, "data")
	assert.Equal(nil, err)
	assert.Equal(shipshape.KeyValueNotFound, kvr)

	// Wrong value.
	kvr, _, err = c.CheckKeyValue(shipshape.KeyValue{
		Key:   "foo.bar[0].baz",
		Value: "zoom",
	}, "data")
	assert.Equal(nil, err)
	assert.Equal(shipshape.KeyValueNotEqual, kvr)

	// Correct value.
	kvr, _, err = c.CheckKeyValue(shipshape.KeyValue{
		Key:   "foo.bar[0].baz",
		Value: "zoo",
	}, "data")
	assert.Equal(nil, err)
	assert.Equal(shipshape.KeyValueEqual, kvr)

	// Optional value not present.
	kvr, _, err = c.CheckKeyValue(shipshape.KeyValue{
		Key:      "foo.bar[0].bazzle",
		Value:    "zoom",
		Optional: true,
	}, "data")
	assert.Equal(nil, err)
	assert.Equal(shipshape.KeyValueEqual, kvr)
}

func TestYamlCheckKeyValueList(t *testing.T) {
	assert := assert.New(t)

	c := shipshape.YamlBase{
		CheckBase: shipshape.CheckBase{
			DataMap: map[string][]byte{
				"data": []byte(`
foo:
  bar:
    - baz
    - zoo
    - zoom

`),
			},
		},
	}
	c.UnmarshalDataMap()

	// Disallowed list not provided.
	_, _, err := c.CheckKeyValue(shipshape.KeyValue{
		Key:    "foo.bar",
		IsList: true,
	}, "data")
	assert.Equal("list of allowed or disallowed values not provided", err.Error())

	var kvr shipshape.KeyValueResult
	var fails []string
	// Disallowed values in yaml.
	kvr, fails, err = c.CheckKeyValue(shipshape.KeyValue{
		Key:        "foo.bar",
		IsList:     true,
		Disallowed: []string{"baz", "zoo"},
	}, "data")
	assert.Equal(nil, err)
	assert.Equal(shipshape.KeyValueDisallowedFound, kvr)
	assert.EqualValues([]string{"baz", "zoo"}, fails)

	// No disallowed values in yaml.
	kvr, fails, _ = c.CheckKeyValue(shipshape.KeyValue{
		Key:        "foo.bar",
		IsList:     true,
		Disallowed: []string{"this should", "be a success"},
	}, "data")
	assert.Equal(shipshape.KeyValueEqual, kvr)
	assert.EqualValues(0, len(fails))

	// Allowed values in yaml all match.
	kvr, fails, _ = c.CheckKeyValue(shipshape.KeyValue{
		Key:     "foo.bar",
		IsList:  true,
		Allowed: []string{"baz", "zoo", "zoom"},
	}, "data")
	assert.Equal(shipshape.KeyValueEqual, kvr)
	assert.EqualValues(0, len(fails))

	// Value not in Allowed list.
	kvr, fails, _ = c.CheckKeyValue(shipshape.KeyValue{
		Key:     "foo.bar",
		IsList:  true,
		Allowed: []string{"baz", "zoo"},
	}, "data")
	assert.Equal(shipshape.KeyValueDisallowedFound, kvr)
	assert.EqualValues([]string{"zoom"}, fails)

}

func TestYamlBase(t *testing.T) {
	assert := assert.New(t)

	c := shipshape.YamlBase{}
	c.HasData(true)
	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.EqualValues([]string{"no data available"}, c.Result.Failures)

	mockCheck := func() shipshape.YamlBase {
		return shipshape.YamlBase{
			CheckBase: shipshape.CheckBase{
				DataMap: map[string][]byte{
					"data": []byte(`
check:
  interval_days: 7
notification:
  emails:
    - admin@example.com
`),
				},
			},
			Values: []shipshape.KeyValue{
				{
					Key:   "check.interval_days",
					Value: "7",
				},
			},
		}
	}

	c = mockCheck()
	c.UnmarshalDataMap()
	c.RunCheck(false)
	assert.Equal(shipshape.Pass, c.Result.Status)
	assert.EqualValues(0, len(c.Result.Failures))
	assert.EqualValues([]string{"[data] 'check.interval_days' equals '7'"}, c.Result.Passes)

	// Wrong key, correct value.
	c = mockCheck()
	c.Values = []shipshape.KeyValue{
		{
			Key:   "check.interval",
			Value: "7",
		},
	}
	c.RunCheck(false)
	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.EqualValues(0, len(c.Result.Passes))
	assert.EqualValues([]string{"[data] 'check.interval' not found"}, c.Result.Failures)

	// Correct key, wrong value.
	c = mockCheck()
	c.Values = []shipshape.KeyValue{
		{
			Key:   "check.interval_days",
			Value: "8",
		},
	}
	c.UnmarshalDataMap()
	c.RunCheck(false)
	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.EqualValues(0, len(c.Result.Passes))
	assert.EqualValues([]string{"[data] 'check.interval_days' equals '7', expected '8'"}, c.Result.Failures)

	// Multiple config values - all correct.
	c = mockCheck()
	c.Values = []shipshape.KeyValue{
		{
			Key:   "check.interval_days",
			Value: "7",
		},
		{
			Key:   "notification.emails[0]",
			Value: "admin@example.com",
		},
	}
	c.UnmarshalDataMap()
	c.RunCheck(false)
	assert.Equal(shipshape.Pass, c.Result.Status)
	assert.EqualValues(0, len(c.Result.Failures))
	assert.EqualValues(
		[]string{
			"[data] 'check.interval_days' equals '7'",
			"[data] 'notification.emails[0]' equals 'admin@example.com'",
		},
		c.Result.Passes)

	// Wildcard key.
	c = mockCheck()
	c.DataMap = map[string][]byte{
		"data": []byte(`
abcd:
  some:
    - thing 1
    - thing 2
    - thing 3
efgh:
  some:
    - thing 1
    - thing 2
    - thing 3
`),
	}
	c.Values = []shipshape.KeyValue{
		{
			Key:        "*.some",
			IsList:     true,
			Disallowed: []string{"thing 2", "thing 4"},
		},
	}
	c.UnmarshalDataMap()
	c.RunCheck(false)
	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.EqualValues(0, len(c.Result.Passes))
	assert.EqualValues([]string{"[data] disallowed *.some: [thing 2]"}, c.Result.Failures)
}

func TestYamlBaseListValues(t *testing.T) {
	assert := assert.New(t)

	mockCheck := func() shipshape.YamlBase {
		return shipshape.YamlBase{
			CheckBase: shipshape.CheckBase{
				DataMap: map[string][]byte{
					"data": []byte(`
foo:
  - a
  - b
  - c
  - d
`),
				},
			},
			Values: []shipshape.KeyValue{
				{
					Key:        "foo",
					IsList:     true,
					Disallowed: []string{"b", "c"},
				},
			},
		}
	}
	c := mockCheck()
	c.UnmarshalDataMap()
	c.RunCheck(false)
	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.EqualValues(0, len(c.Result.Passes))
	assert.EqualValues([]string{"[data] disallowed foo: [b, c]"}, c.Result.Failures)

	c = mockCheck()
	c.Values[0].Disallowed = []string{"e"}
	c.UnmarshalDataMap()
	c.RunCheck(false)
	assert.Equal(shipshape.Pass, c.Result.Status)
	assert.EqualValues(0, len(c.Result.Failures))
	assert.EqualValues([]string{"[data] no disallowed 'foo'"}, c.Result.Passes)
}

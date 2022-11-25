package phpstan_test

import (
	"os/exec"
	"reflect"
	"testing"

	"github.com/salsadigitalauorg/shipshape/internal"
	"github.com/salsadigitalauorg/shipshape/pkg/phpstan"
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/stretchr/testify/assert"
)

func TestExecCommandHelper(t *testing.T) {
	internal.TestExecCommandHelper(t)
}

func TestRegisterChecks(t *testing.T) {
	checksMap := map[shipshape.CheckType]string{
		phpstan.PhpStan: "*phpstan.PhpStanCheck",
	}
	for ct, ts := range checksMap {
		c := shipshape.ChecksRegistry[ct]()
		ctype := reflect.TypeOf(c).String()
		if ctype != ts {
			t.Errorf("expecting check of type '%s', got '%s'", ts, ctype)
		}
	}
}

func TestMerge(t *testing.T) {
	assert := assert.New(t)

	c := phpstan.PhpStanCheck{
		CheckBase: shipshape.CheckBase{Name: "phpstancheck1"},
		Bin:       "/path/to/phpstan",
		Config:    "/path/to/config",
		Paths:     []string{"path1", "path2"},
	}
	err := c.Merge(&phpstan.PhpStanCheck{
		Bin: "/new/path/to/phpstan",
	})
	assert.Nil(err)
	assert.EqualValues(phpstan.PhpStanCheck{
		CheckBase: shipshape.CheckBase{Name: "phpstancheck1"},
		Bin:       "/new/path/to/phpstan",
		Config:    "/path/to/config",
		Paths:     []string{"path1", "path2"},
	}, c)

	err = c.Merge(&phpstan.PhpStanCheck{
		Config: "/path/to/new/config",
	})
	assert.Nil(err)
	assert.EqualValues(phpstan.PhpStanCheck{
		CheckBase: shipshape.CheckBase{Name: "phpstancheck1"},
		Bin:       "/new/path/to/phpstan",
		Config:    "/path/to/new/config",
		Paths:     []string{"path1", "path2"},
	}, c)

	err = c.Merge(&phpstan.PhpStanCheck{
		Paths: []string{"path3", "path4"},
	})
	assert.Nil(err)
	assert.EqualValues(phpstan.PhpStanCheck{
		CheckBase: shipshape.CheckBase{Name: "phpstancheck1"},
		Bin:       "/new/path/to/phpstan",
		Config:    "/path/to/new/config",
		Paths:     []string{"path3", "path4"},
	}, c)

	err = c.Merge(&phpstan.PhpStanCheck{
		CheckBase: shipshape.CheckBase{Name: "phpstancheck2"},
		Bin:       "/some/other/path/to/phpstan",
	})
	assert.Error(err, "can only merge checks with the same name")
}

func TestBinPathProvided(t *testing.T) {
	assert := assert.New(t)
	c := phpstan.PhpStanCheck{
		Bin:    "/my/custom/path/phpstan",
		Config: "/path/to/config",
	}

	assert.Equal("/my/custom/path/phpstan", c.GetBinary())
}

func TestBinPathDefault(t *testing.T) {
	assert := assert.New(t)
	c := phpstan.PhpStanCheck{
		Config: "/path/to/config",
	}

	assert.Equal("vendor/phpstan/phpstan/phpstan", c.GetBinary())
}

func TestFetchDataBinNotExists(t *testing.T) {
	assert := assert.New(t)
	phpstan.ExecCommand = internal.FakeExecCommand
	internal.MockedExitStatus = 2
	internal.MockedStderr = "/my/custom/path/phpstan: no such file or directory"
	defer func() {
		phpstan.ExecCommand = exec.Command
		internal.MockedExitStatus = 0
		internal.MockedStderr = ""
	}()

	// Command not found.
	c := phpstan.PhpStanCheck{
		Bin:    "/my/custom/path/phpstan",
		Config: "/path/to/config",
		Paths:  []string{"/some/path/to/analyse"},
	}
	c.FetchData()

	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.EqualValues([]string{"Phpstan failed to run: /my/custom/path/phpstan: no such file or directory"}, c.Result.Failures)
}

func TestFetchDataBinExists(t *testing.T) {
	assert := assert.New(t)
	phpstan.ExecCommand = internal.FakeExecCommand
	internal.MockedStdout = `{"totals":{"errors":0,"file_errors":0},"files":[],"errors":[]}`
	defer func() {
		phpstan.ExecCommand = exec.Command
		internal.MockedStdout = ""
	}()

	c := phpstan.PhpStanCheck{
		Bin:    "/my/custom/path/phpstan",
		Config: "path/to/config",
		Paths:  []string{"relative/path/to/analyse"},
	}
	c.FetchData()

	assert.NotEqual(shipshape.Pass, c.Result.Status)
	assert.NotEqual(shipshape.Fail, c.Result.Status)
	assert.Equal(internal.MockedStdout, string(c.DataMap["phpstan"]))
}

func TestUnmarshalDataMap(t *testing.T) {
	assert := assert.New(t)
	// No DataMap.
	c := phpstan.PhpStanCheck{}
	c.UnmarshalDataMap()
	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.EqualValues([]string{"no data provided"}, c.Result.Failures)

	// Empty data.
	c = phpstan.PhpStanCheck{
		CheckBase: shipshape.CheckBase{
			DataMap: map[string][]byte{
				"phpstan": []byte(`{"totals":{"errors":0,"file_errors":0},"files":[],"errors":[]}`),
			},
		},
	}
	c.UnmarshalDataMap()
	assert.NotEqual(shipshape.Pass, c.Result.Status)
	assert.NotEqual(shipshape.Fail, c.Result.Status)
	filesRaw := reflect.ValueOf(c).FieldByName("phpstanResult").FieldByName("FilesRaw")
	assert.Equal("[]", string(filesRaw.Bytes()))

	// Invalid files data.
	c = phpstan.PhpStanCheck{
		CheckBase: shipshape.CheckBase{
			DataMap: map[string][]byte{
				"phpstan": []byte(`{"files":["foo"]}`),
			},
		},
	}
	c.UnmarshalDataMap()
	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.Contains(c.Result.Failures[0], "json: cannot unmarshal array into Go value of type map[string]struct")
}

func TestRunCheck(t *testing.T) {
	assert := assert.New(t)

	// No file errors.
	c := phpstan.PhpStanCheck{
		CheckBase: shipshape.CheckBase{
			DataMap: map[string][]byte{
				"phpstan": []byte(`{"totals":{"errors":0,"file_errors":0},"files":[],"errors":[]}`),
			},
		},
	}
	c.UnmarshalDataMap()
	c.RunCheck()
	assert.Equal(shipshape.Pass, c.Result.Status)
	assert.EqualValues([]string{"no error found"}, c.Result.Passes)

	// PHP errors detected.
	c = phpstan.PhpStanCheck{
		CheckBase: shipshape.CheckBase{
			DataMap: map[string][]byte{
				"phpstan": []byte(`{"totals":{"errors":0,"file_errors":1},"files":{"/app/web/themes/custom/custom/test-theme/info.php":{"errors":1,"messages":[{"message":"Calling curl_exec() is forbidden, please change the code","line":3,"ignorable":true}]}},"errors":[]}`),
			},
		},
	}
	c.UnmarshalDataMap()
	c.RunCheck()
	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.EqualValues([]string{"[/app/web/themes/custom/custom/test-theme/info.php] Line 3: Calling curl_exec() is forbidden, please change the code"}, c.Result.Failures)

	// Other errors found in files.
	c = phpstan.PhpStanCheck{
		CheckBase: shipshape.CheckBase{
			DataMap: map[string][]byte{
				"phpstan": []byte(`{"totals":{"errors":0,"file_errors":1},"files":[],"errors":["Error found in file foo"]}`),
			},
		},
	}
	c.UnmarshalDataMap()
	c.RunCheck()
	assert.Equal(shipshape.Fail, c.Result.Status)
	assert.EqualValues([]string{"Error found in file foo"}, c.Result.Failures)
}

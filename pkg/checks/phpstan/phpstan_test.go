package phpstan_test

import (
	"os"
	"os/exec"
	"reflect"
	"testing"

	. "github.com/salsadigitalauorg/shipshape/pkg/checks/phpstan"
	"github.com/salsadigitalauorg/shipshape/pkg/command"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/internal"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
	"github.com/stretchr/testify/assert"
)

func TestRegisterChecks(t *testing.T) {
	checksMap := map[config.CheckType]string{
		PhpStan: "*phpstan.PhpStanCheck",
	}
	for ct, ts := range checksMap {
		c := config.ChecksRegistry[ct]()
		ctype := reflect.TypeOf(c).String()
		if ctype != ts {
			t.Errorf("expecting check of type '%s', got '%s'", ts, ctype)
		}
	}
}

func TestMerge(t *testing.T) {
	assert := assert.New(t)

	c := PhpStanCheck{
		CheckBase: config.CheckBase{Name: "phpstancheck1"},
		Bin:       "/path/to/phpstan",
		Config:    "/path/to/config",
		Paths:     []string{"path1", "path2"},
	}
	err := c.Merge(&PhpStanCheck{
		Bin: "/new/path/to/phpstan",
	})
	assert.Nil(err)
	assert.EqualValues(PhpStanCheck{
		CheckBase: config.CheckBase{Name: "phpstancheck1"},
		Bin:       "/new/path/to/phpstan",
		Config:    "/path/to/config",
		Paths:     []string{"path1", "path2"},
	}, c)

	err = c.Merge(&PhpStanCheck{
		Config: "/path/to/new/config",
	})
	assert.Nil(err)
	assert.EqualValues(PhpStanCheck{
		CheckBase: config.CheckBase{Name: "phpstancheck1"},
		Bin:       "/new/path/to/phpstan",
		Config:    "/path/to/new/config",
		Paths:     []string{"path1", "path2"},
	}, c)

	err = c.Merge(&PhpStanCheck{
		Paths: []string{"path3", "path4"},
	})
	assert.Nil(err)
	assert.EqualValues(PhpStanCheck{
		CheckBase: config.CheckBase{Name: "phpstancheck1"},
		Bin:       "/new/path/to/phpstan",
		Config:    "/path/to/new/config",
		Paths:     []string{"path3", "path4"},
	}, c)

	err = c.Merge(&PhpStanCheck{
		CheckBase: config.CheckBase{Name: "phpstancheck2"},
		Bin:       "/some/other/path/to/phpstan",
	})
	assert.Error(err, "can only merge checks with the same name")
}

func TestBinPathProvided(t *testing.T) {
	assert := assert.New(t)
	c := PhpStanCheck{
		Bin:    "/my/custom/path/phpstan",
		Config: "/path/to/config",
	}

	assert.Equal("/my/custom/path/phpstan", c.GetBinary())
}

func TestBinPathDefault(t *testing.T) {
	assert := assert.New(t)
	c := PhpStanCheck{
		Config: "/path/to/config",
	}

	assert.Equal("vendor/phpstan/phpstan/phpstan", c.GetBinary())
}

func TestFetchDataPathNotExists(t *testing.T) {
	assert := assert.New(t)
	// No files found to analyse.
	c := PhpStanCheck{
		Paths: []string{
			"/path/to/non/existing",
			"other/non/existing",
		},
	}
	c.FetchData()

	assert.Equal(result.Pass, c.Result.Status)
	assert.EqualValues([]string{"no paths found to run phpstan on"}, c.Result.Passes)
}

func TestFetchDataBinNotExists(t *testing.T) {
	assert := assert.New(t)

	curShellCommander := command.ShellCommander
	defer func() { command.ShellCommander = curShellCommander }()
	command.ShellCommander = internal.ShellCommanderMaker(
		nil,
		&exec.ExitError{Stderr: []byte("/my/custom/path/phpstan: no such file or directory")},
		nil)

	// Command not found.
	dir, _ := os.Getwd()
	c := PhpStanCheck{
		Bin:    "/my/custom/path/phpstan",
		Config: "/path/to/config",
		Paths:  []string{dir},
	}
	c.FetchData()

	assert.Equal(result.Fail, c.Result.Status)
	assert.EqualValues([]string{"Phpstan failed to run: /my/custom/path/phpstan: no such file or directory"}, c.Result.Failures)
}

func TestFetchDataBinExists(t *testing.T) {
	assert := assert.New(t)

	expectedStdout := `{"totals":{"errors":0,"file_errors":0},"files":[],"errors":[]}`

	curShellCommander := command.ShellCommander
	defer func() { command.ShellCommander = curShellCommander }()
	command.ShellCommander = internal.ShellCommanderMaker(&expectedStdout, nil, nil)

	dir, _ := os.Getwd()
	c := PhpStanCheck{
		Bin:    "/my/custom/path/phpstan",
		Config: "path/to/config",
		Paths:  []string{dir},
	}
	c.FetchData()

	assert.NotEqual(result.Pass, c.Result.Status)
	assert.NotEqual(result.Fail, c.Result.Status)
	assert.Equal([]byte(expectedStdout), c.DataMap["phpstan"])
}

func TestHasData(t *testing.T) {
	t.Run("no data, ignore failures", func(t *testing.T) {
		assert := assert.New(t)
		c := PhpStanCheck{}
		assert.False(c.HasData(false))
		assert.NotEqual(result.Pass, c.Result.Status)
		assert.NotEqual(result.Fail, c.Result.Status)
	})

	t.Run("no data, with fail", func(t *testing.T) {
		assert := assert.New(t)
		c := PhpStanCheck{}
		assert.False(c.HasData(true))
		assert.Equal(result.Fail, c.Result.Status)
		assert.EqualValues([]string{"no data available"}, c.Result.Failures)
	})

	t.Run("no data, but passed", func(t *testing.T) {
		assert := assert.New(t)
		c := PhpStanCheck{}
		c.AddPass("passed")
		assert.True(c.HasData(true))
		assert.NotEqual(result.Pass, c.Result.Status)
		assert.NotEqual(result.Fail, c.Result.Status)
		assert.EqualValues([]string{"passed"}, c.Result.Passes)
	})
}

func TestUnmarshalDataMap(t *testing.T) {
	assert := assert.New(t)
	// No DataMap.
	c := PhpStanCheck{}
	c.UnmarshalDataMap()
	assert.Equal(result.Pass, c.Result.Status)
	assert.EqualValues([]string{"Unhandled PHPStan response, unable to determine status."}, c.Result.Warnings)

	// Empty data.
	c = PhpStanCheck{
		CheckBase: config.CheckBase{
			DataMap: map[string][]byte{
				"phpstan": []byte(`{"totals":{"errors":0,"file_errors":0},"files":[],"errors":[]}`),
			},
		},
	}
	c.UnmarshalDataMap()
	assert.NotEqual(result.Pass, c.Result.Status)
	assert.NotEqual(result.Fail, c.Result.Status)
	filesRaw := reflect.ValueOf(c).FieldByName("phpstanResult").FieldByName("FilesRaw")
	assert.Equal("[]", string(filesRaw.Bytes()))

	// Invalid files data.
	c = PhpStanCheck{
		CheckBase: config.CheckBase{
			DataMap: map[string][]byte{
				"phpstan": []byte(`{"files":["foo"]}`),
			},
		},
	}
	c.UnmarshalDataMap()
	assert.Equal(result.Fail, c.Result.Status)
	assert.Contains(c.Result.Failures[0], "json: cannot unmarshal array into Go value of type map[string]struct")
}

func TestRunCheck(t *testing.T) {
	assert := assert.New(t)

	// No file errors.
	c := PhpStanCheck{
		CheckBase: config.CheckBase{
			DataMap: map[string][]byte{
				"phpstan": []byte(`{"totals":{"errors":0,"file_errors":0},"files":[],"errors":[]}`),
			},
		},
	}
	c.UnmarshalDataMap()
	c.RunCheck()
	assert.Equal(result.Pass, c.Result.Status)
	assert.EqualValues([]string{"no error found"}, c.Result.Passes)

	// PHP errors detected.
	c = PhpStanCheck{
		CheckBase: config.CheckBase{
			DataMap: map[string][]byte{
				"phpstan": []byte(`{"totals":{"errors":0,"file_errors":1},"files":{"/app/web/themes/custom/custom/test-theme/info.php":{"errors":1,"messages":[{"message":"Calling curl_exec() is forbidden, please change the code","line":3,"ignorable":true}]}},"errors":[]}`),
			},
		},
	}
	c.UnmarshalDataMap()
	c.RunCheck()
	assert.Equal(result.Fail, c.Result.Status)
	assert.EqualValues([]string{"[/app/web/themes/custom/custom/test-theme/info.php] Line 3: Calling curl_exec() is forbidden, please change the code"}, c.Result.Failures)

	// Other errors found in files.
	c = PhpStanCheck{
		CheckBase: config.CheckBase{
			DataMap: map[string][]byte{
				"phpstan": []byte(`{"totals":{"errors":0,"file_errors":1},"files":[],"errors":["Error found in file foo"]}`),
			},
		},
	}
	c.UnmarshalDataMap()
	c.RunCheck()
	assert.Equal(result.Fail, c.Result.Status)
	assert.EqualValues([]string{"Error found in file foo"}, c.Result.Failures)
}

func TestInvalidOutput(t *testing.T) {
	assert := assert.New(t)

	c := PhpStanCheck{
		CheckBase: config.CheckBase{
			DataMap: map[string][]byte{
				"phpstan": []byte(""),
			},
		},
	}

	c.UnmarshalDataMap()

	assert.Equal(c.Result.Status, result.Pass)
	assert.Equal(c.GetResult().Warnings, []string{
		"Unhandled PHPStan response, unable to determine status.",
	})
}

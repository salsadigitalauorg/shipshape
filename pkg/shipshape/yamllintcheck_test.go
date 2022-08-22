package shipshape_test

import (
	"testing"

	"github.com/salsadigitalauorg/shipshape/internal"
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
)

func TestYamlLintCheck(t *testing.T) {
	mockCheck := func(file string, files []string, ignoreMissing bool) shipshape.YamlLintCheck {
		return shipshape.YamlLintCheck{
			YamlCheck: shipshape.YamlCheck{
				YamlBase: shipshape.YamlBase{
					CheckBase: shipshape.CheckBase{
						Name:    "Test yaml lint",
						DataMap: map[string][]byte{},
					},
				},
				File:          file,
				Files:         files,
				IgnoreMissing: ignoreMissing,
			},
		}
	}

	c := mockCheck("", []string{}, false)
	c.Init("testdata", shipshape.YamlLint)
	c.FetchData()
	if msg, ok := internal.EnsureFail(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{"no file provided"}); !ok {
		t.Error(msg)
	}

	c = mockCheck("non-existent-file.yml", []string{}, true)
	c.Init("testdata", shipshape.YamlLint)
	c.FetchData()
	if msg, ok := internal.EnsureNoFail(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string{"File testdata/non-existent-file.yml does not exist"}); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}

	c = mockCheck("", []string{"non-existent-file.yml", "yaml-invalid.yml"}, true)
	c.Init("testdata", shipshape.YamlLint)
	c.FetchData()
	if msg, ok := internal.EnsureNoFail(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string{
		"File testdata/non-existent-file.yml does not exist",
		"File testdata/yaml-invalid.yml does not exist",
	}); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}

	c = mockCheck("non-existent-file.yml", []string{}, false)
	c.Init("testdata", shipshape.YamlLint)
	c.FetchData()
	if msg, ok := internal.EnsureFail(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{"open testdata/non-existent-file.yml: no such file or directory"}); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}

	c = mockCheck("", []string{"non-existent-file.yml", "yamllint-invalid.yml"}, false)
	c.Init("testdata", shipshape.YamlLint)
	c.FetchData()
	if msg, ok := internal.EnsureFail(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{"open testdata/non-existent-file.yml: no such file or directory"}); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}

	c = mockCheck("", []string{}, false)
	c.Init("testdata", shipshape.YamlLint)
	c.DataMap["yaml-invalid.yml"] = []byte(`
this: is invalid
this: yaml
`)
	c.UnmarshalDataMap()
	if msg, ok := internal.EnsureFail(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{"[yaml-invalid.yml] line 3: mapping key \"this\" already defined at line 2"}); !ok {
		t.Error(msg)
	}

	c = mockCheck("", []string{}, false)
	c.Init("testdata", shipshape.YamlLint)
	c.DataMap["yaml-valid.yml"] = []byte(`
this: is
valid: yaml
`)
	c.UnmarshalDataMap()
	if msg, ok := internal.EnsurePass(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string{"yaml-valid.yml has valid yaml."}); !ok {
		t.Error(msg)
	}
}

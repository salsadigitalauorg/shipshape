package shipshape_test

import (
	"testing"

	"github.com/salsadigitalauorg/shipshape/internal"
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
)

func TestYamlCheck(t *testing.T) {
	mockCheck := func() shipshape.YamlCheck {
		return shipshape.YamlCheck{
			YamlBase: shipshape.YamlBase{
				Values: []shipshape.KeyValue{
					{
						Key:   "check.interval_days",
						Value: "7",
					},
				},
			},
			Path: "yaml",
		}
	}

	c := mockCheck()
	c.FetchData()
	if _, ok := internal.EnsureFail(t, &c.CheckBase); !ok {
		t.Error("Check with no File or Pattern should Fail")
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{"no file provided"}); !ok {
		t.Error(msg)
	}

	// Non-existent file.
	c = mockCheck()
	c.Init("testdata", shipshape.Yaml)
	c.File = "non-existent.yml"
	c.FetchData()
	if _, ok := internal.EnsureFail(t, &c.CheckBase); !ok {
		t.Error("Check with non-existent file should Fail")
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{"open testdata/yaml/non-existent.yml: no such file or directory"}); !ok {
		t.Error(msg)
	}

	// Non-existent file with ignore missing.
	c = mockCheck()
	c.File = "non-existent.yml"
	c.IgnoreMissing = true
	c.FetchData()
	if _, ok := internal.EnsurePass(t, &c.CheckBase); !ok {
		t.Error("Check with non-existent file when ignoring missing should Pass already.")
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string{"File testdata/yaml/non-existent.yml does not exist"}); !ok {
		t.Error(msg)
	}

	// Single file.
	c = mockCheck()
	c.File = "update.settings.yml"
	c.FetchData()
	if msg, ok := internal.EnsureNoFail(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if !c.HasData(false) {
		t.Errorf("c.DataMap should be filled, but is empty.")
	}
	c.UnmarshalDataMap()
	c.RunCheck()
	if msg, ok := internal.EnsurePass(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string{"[yaml/update.settings.yml] 'check.interval_days' equals '7'"}); !ok {
		t.Error(msg)
	}

	// Bad File pattern.
	c = mockCheck()
	c.Pattern = "*.bar.yml"
	c.Path = ""
	c.FetchData()
	if _, ok := internal.EnsureFail(t, &c.CheckBase); !ok {
		t.Error("Check with bad file pattern should fail")
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{"error parsing regexp: missing argument to repetition operator: `*`"}); !ok {
		t.Error(msg)
	}

	// File pattern with no matching files.
	c = mockCheck()
	c.Pattern = "bla.*.yml"
	c.FetchData()
	if msg, ok := internal.EnsureFail(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{"no matching config files found"}); !ok {
		t.Error(msg)
	}

	// File pattern with no matching files, ignoring missing.
	c = mockCheck()
	c.Pattern = "bla.*.yml"
	c.IgnoreMissing = true
	c.FetchData()
	if _, ok := internal.EnsurePass(t, &c.CheckBase); !ok {
		t.Error("Check with non-existent file pattern when ignoring missing should Pass already.")
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string{"no matching config files found"}); !ok {
		t.Error(msg)
	}

	// Correct single file pattern & value.
	c = mockCheck()
	c.Pattern = "foo.bar.yml"
	c.Path = "yaml/dir/subdir"
	c.FetchData()
	if c.Result.Status == shipshape.Fail {
		t.Error("Check should not Fail yet")
	}
	if len(c.Result.Failures) > 0 {
		t.Errorf("there should be no Failure, got: %#v", c.Result.Failures)
	}
	c.UnmarshalDataMap()
	c.RunCheck()
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string{"[testdata/yaml/dir/subdir/foo.bar.yml] 'check.interval_days' equals '7'"}); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}

	// Recursive file lookup.
	c = mockCheck()
	c.Pattern = ".*.bar.yml"
	c.FetchData()
	if c.Result.Status == shipshape.Fail {
		t.Error("Check should not Fail yet")
	}
	if len(c.Result.Failures) > 0 {
		t.Errorf("there should be no Failure, got: %#v", c.Result.Failures)
	}
	c.UnmarshalDataMap()
	c.RunCheck()
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string{"[testdata/yaml/dir/foo.bar.yml] 'check.interval_days' equals '7'", "[testdata/yaml/dir/subdir/foo.bar.yml] 'check.interval_days' equals '7'", "[testdata/yaml/foo.bar.yml] 'check.interval_days' equals '7'"}); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{"[testdata/yaml/dir/subdir/zoom.bar.yml] 'check.interval_days' equals '5'", "[testdata/yaml/dir/zoom.bar.yml] 'check.interval_days' equals '5'", "[testdata/yaml/zoom.bar.yml] 'check.interval_days' equals '5'"}); !ok {
		t.Error(msg)
	}
}

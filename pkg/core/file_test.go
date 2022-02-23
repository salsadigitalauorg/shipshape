package core_test

import (
	"salsadigitalauorg/shipshape/internal"
	"salsadigitalauorg/shipshape/pkg/core"
	"testing"
)

func TestFileCheck(t *testing.T) {
	c := core.FileCheck{
		Path:              "file-non-existent",
		DisallowedPattern: "^(adminer|phpmyadmin|bigdump)?\\.php$",
	}
	c.Init("testdata", core.File)
	c.RunCheck()
	if msg, ok := internal.EnsureFail(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{"lstat testdata/file-non-existent: no such file or directory"}); !ok {
		t.Error(msg)
	}

	c = core.FileCheck{
		Path:              "file",
		DisallowedPattern: "^(adminer|phpmyadmin|bigdump)?\\.php$",
	}
	c.Init("testdata", core.File)
	c.RunCheck()
	if msg, ok := internal.EnsureFail(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{
		"Illegal file found: testdata/file/adminer.php",
		"Illegal file found: testdata/file/sub/phpmyadmin.php",
	}); !ok {
		t.Error(msg)
	}

	c = core.FileCheck{
		Path:              "file/correct",
		DisallowedPattern: "^(adminer|phpmyadmin|bigdump)?\\.php$",
	}
	c.Init("testdata", core.File)
	c.RunCheck()
	if msg, ok := internal.EnsurePass(t, &c.CheckBase); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string{"No illegal files"}); !ok {
		t.Error(msg)
	}

}

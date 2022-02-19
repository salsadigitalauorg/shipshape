package core_test

import (
	"reflect"
	"salsadigitalauorg/shipshape/pkg/core"
	"testing"
)

func TestFileCheck(t *testing.T) {
	c := core.FileCheck{
		CheckBase: core.CheckBase{
			ProjectDir: "testdata/file-non-existent",
		},
		DisallowedPattern: "^(adminer|phpmyadmin|bigdump)?\\.php$",
	}
	c.RunCheck()
	if c.Result.Status != core.Fail {
		t.Error("Check status should be Fail")
	}
	if len(c.Result.Passes) > 0 {
		t.Errorf("There should be no Passes, got %+v", c.Result.Passes)
	}
	if len(c.Result.Failures) != 1 {
		t.Errorf("There should be exactly 1 Failure, got %+v", c.Result.Failures)
	}
	if c.Result.Failures[0] != "lstat testdata/file-non-existent: no such file or directory" {
		t.Errorf("Pass message should be 'lstat testdata/file-non-existent: no such file or directory', got %s", c.Result.Failures[0])
	}

	c = core.FileCheck{
		CheckBase: core.CheckBase{
			ProjectDir: "testdata/file",
		},
		DisallowedPattern: "^(adminer|phpmyadmin|bigdump)?\\.php$",
	}
	c.RunCheck()
	if c.Result.Status != core.Fail {
		t.Error("Check status should be Fail")
	}
	if len(c.Result.Passes) > 0 {
		t.Errorf("There should be no Passes, got %+v", c.Result.Passes)
	}
	if len(c.Result.Failures) != 2 {
		t.Errorf("There should be exactly 2 Failures, got %+v", c.Result.Failures)
	}
	expectedFailures := []string{
		"Illegal file found: testdata/file/adminer.php",
		"Illegal file found: testdata/file/sub/phpmyadmin.php",
	}
	if !reflect.DeepEqual(c.Result.Failures, expectedFailures) {
		t.Errorf("Values for Failures are not as expected, got %+v", c.Result.Failures)
	}

	c = core.FileCheck{
		CheckBase: core.CheckBase{
			ProjectDir: "testdata/file/correct",
		},
		DisallowedPattern: "^(adminer|phpmyadmin|bigdump)?\\.php$",
	}
	c.RunCheck()
	if c.Result.Status != core.Pass {
		t.Error("Check status should be Pass")
	}
	if len(c.Result.Failures) > 0 {
		t.Errorf("There should be no Failures, got %+v", c.Result.Failures)
	}
	if len(c.Result.Passes) != 1 {
		t.Errorf("There should be exactly 1 Pass, got %+v", c.Result.Passes)
	}
	if c.Result.Passes[0] != "No illegal files" {
		t.Errorf("Pass message should be 'No illegal files', got %s", c.Result.Passes[0])
	}

}

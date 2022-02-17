package core_test

import (
	"reflect"
	"salsadigitalauorg/shipshape/pkg/core"
	"testing"
)

func TestFileCheck(t *testing.T) {
	c := core.FileCheck{
		CheckBase: core.CheckBase{
			ProjectDir: "testdata/file",
		},
		DisallowedPattern: "^(adminer|phpmyadmin|bigdump)?\\.php$",
	}
	c.RunCheck()
	if c.Result.Status != core.Fail {
		t.Errorf("Check status should be Fail, got %s", c.Result.Status)
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
}

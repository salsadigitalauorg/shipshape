package utils_test

import (
	"reflect"
	"salsadigitalauorg/shipshape/pkg/utils"
	"testing"
)

func TestFindFiles(t *testing.T) {
	_, err := utils.FindFiles("", "", "")
	if err == nil || err.Error() != "directory not provided" {
		t.Error("empty directory should fail.")
	}

	_, err = utils.FindFiles("testdata/findfiles", "", "")
	if err == nil || err.Error() != "pattern not provided" {
		t.Error("empty pattern should fail.")
	}

	files, err := utils.FindFiles("testdata/findfiles", "user.role.*.yml", "")
	if err != nil {
		t.Errorf("FindFiles should succeed, but failed: %+v", err)
	}
	expectedFiles := []string{
		"testdata/findfiles/a/b/user.role.bogus.yml",
		"testdata/findfiles/user.role.admin.yml",
		"testdata/findfiles/user.role.author.yml",
		"testdata/findfiles/user.role.editor.yml",
	}
	if len(files) != 4 || !reflect.DeepEqual(files, expectedFiles) {
		t.Errorf("There should be exactly 4 files, got: %+v", files)
	}

	files, err = utils.FindFiles("testdata/findfiles", "^user\\.role\\..*\\.yml$", "user.role.author.yml")
	if err != nil {
		t.Errorf("FindFiles should succeed, but failed: %+v", err)
	}
	expectedFiles = []string{
		"testdata/findfiles/a/b/user.role.bogus.yml",
		"testdata/findfiles/user.role.admin.yml",
		"testdata/findfiles/user.role.editor.yml",
	}
	if len(files) != 3 || !reflect.DeepEqual(files, expectedFiles) {
		t.Errorf("There should be exactly 3 files, got: %+v", files)
	}
}

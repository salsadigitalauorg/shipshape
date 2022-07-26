package utils_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/utils"
	"github.com/stretchr/testify/assert"
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

func TestStringSliceContains(t *testing.T) {
	assert := assert.New(t)
	assert.False(utils.StringSliceContains([]string{}, "foo"))
	assert.False(utils.StringSliceContains([]string{"bar"}, "foo"))
	assert.True(utils.StringSliceContains([]string{"bar", "foo"}, "foo"))
}

func TestIntSliceContains(t *testing.T) {
	assert := assert.New(t)
	assert.False(utils.IntSliceContains([]int{}, 10))
	assert.False(utils.IntSliceContains([]int{5}, 10))
	assert.True(utils.IntSliceContains([]int{5, 10}, 10))
}

func TestStringSlicesIntersect(t *testing.T) {
	intersect := utils.StringSlicesIntersect(
		[]string{"foo"}, []string{})
	if len(intersect) != 0 {
		t.Errorf("Intersect should be empty, got '%+v'", intersect)
	}

	intersect = utils.StringSlicesIntersect(
		[]string{"foo"}, []string{"bar"})
	if len(intersect) != 0 {
		t.Errorf("Intersect should be empty, got '%+v'", intersect)
	}

	intersect = utils.StringSlicesIntersect(
		[]string{"foo"}, []string{"bar", "foo"})
	expectedIntersect := []string{"foo"}
	if len(intersect) != 1 || !reflect.DeepEqual(intersect, expectedIntersect) {
		t.Errorf("Intersect should have 1 item, got '%+v'", intersect)
	}

	intersect = utils.StringSlicesIntersect(
		[]string{"foo", "baz", "zoom"}, []string{"bar", "foo", "zoo", "zoom"})
	expectedIntersect = []string{"foo", "zoom"}
	if len(intersect) != 2 || !reflect.DeepEqual(intersect, expectedIntersect) {
		t.Errorf("Intersect should have 2 item, got '%+v'", intersect)
	}
}

func TestStringIsUrl(t *testing.T) {
	isUrl := utils.StringIsUrl("foo/bar.yml")
	if isUrl {
		t.Error("expected isUrl to be false, got true")
	}

	isUrl = utils.StringIsUrl("~/foo/bar.yml")
	if isUrl {
		t.Error("expected isUrl to be false, got true")
	}

	isUrl = utils.StringIsUrl("/home/user/foo/bar.yml")
	if isUrl {
		t.Error("expected isUrl to be false, got true")
	}

	isUrl = utils.StringIsUrl("https://example.com/foo.yml")
	if !isUrl {
		t.Error("expected isUrl to be true, got false")
	}

	isUrl = utils.StringIsUrl("https://127.0.0.1:8080/foo.yml")
	if !isUrl {
		t.Error("expected isUrl to be true, got false")
	}
}

func TestFetchContentFromUrl(t *testing.T) {
	expected := "dummy data"
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, expected)
	}))
	defer svr.Close()

	c, err := utils.FetchContentFromUrl(svr.URL + "/foo.yml")
	if err != nil {
		t.Errorf("expected err to be nil got %v", err)
	}
	if string(c) != expected {
		t.Errorf("expected content to be %s, got %s", expected, c)
	}
}

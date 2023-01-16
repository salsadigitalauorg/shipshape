package utils_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	. "github.com/salsadigitalauorg/shipshape/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestIsFileInDirs(t *testing.T) {
	assert := assert.New(t)

	t.Run("fileAbsent", func(t *testing.T) {
		assert.False(IsFileInDirs(
			"testdata/isfileindirs",
			"testdata/isfileindirs/dir1/dir2/file1",
			[]string{
				"dir3/dir4",
				"dir5/dir6",
				"dir7",
			}))
		assert.False(IsFileInDirs(
			"testdata/isfileindirs",
			"testdata/isfileindirs/dir1/dir2/file2",
			[]string{
				"dir2",
				"dir5/dir6",
				"dir7",
			}))
	})

	t.Run("filePresent", func(t *testing.T) {
		assert.True(IsFileInDirs(
			"testdata/isfileindirs",
			"testdata/isfileindirs/dir1/dir2/file1",
			[]string{
				"dir1",
				"dir5/dir6",
				"dir7",
			}))
		assert.True(IsFileInDirs(
			"testdata/isfileindirs",
			"testdata/isfileindirs/dir1/dir2/file2",
			[]string{
				"dir1/dir2",
				"dir5/dir6",
				"dir7",
			}))
		assert.True(IsFileInDirs(
			"testdata/isfileindirs",
			"testdata/isfileindirs/dir2/dir1/file4",
			[]string{
				"dir1/dir2",
				"dir2",
				"dir5/dir6",
				"dir7",
			}))
	})
}

func TestFindFiles(t *testing.T) {
	assert := assert.New(t)

	t.Run("missingArgs", func(t *testing.T) {
		_, err := FindFiles("", "", "", nil)
		assert.Error(errors.New("directory not provided"), err)

		_, err = FindFiles("testdata/findfiles", "", "", nil)
		assert.Error(errors.New("pattern not provided"), err)
	})

	t.Run("simpleFilePattern", func(t *testing.T) {
		files, err := FindFiles("testdata/findfiles", "user.role.*.yml", "", nil)
		assert.NoError(err)
		assert.ElementsMatch([]string{
			"testdata/findfiles/a/b/user.role.bogus.yml",
			"testdata/findfiles/user.role.admin.yml",
			"testdata/findfiles/user.role.author.yml",
			"testdata/findfiles/user.role.editor.yml",
		}, files)
	})

	t.Run("filePatternWithExclusion", func(t *testing.T) {
		files, err := FindFiles("testdata/findfiles",
			"^user\\.role\\..*\\.yml$", "user.role.author.yml", nil)
		assert.NoError(err)
		assert.ElementsMatch([]string{
			"testdata/findfiles/a/b/user.role.bogus.yml",
			"testdata/findfiles/user.role.admin.yml",
			"testdata/findfiles/user.role.editor.yml",
		}, files)
	})

	t.Run("dirPattern", func(t *testing.T) {
		files, err := FindFiles("testdata/findfiles", ".*.yml", "", nil)
		assert.NoError(err)
		assert.ElementsMatch([]string{
			"testdata/findfiles/a/some-file.yml",
			"testdata/findfiles/a/b/user.role.bogus.yml",
			"testdata/findfiles/user.admin.yml",
			"testdata/findfiles/user.role.admin.yml",
			"testdata/findfiles/user.role.author.yml",
			"testdata/findfiles/user.role.editor.yml",
			"testdata/findfiles/node_modules/yamllint-invalid.yml",
		}, files)
	})

	t.Run("dirPatternWithExclusion", func(t *testing.T) {
		files, err := FindFiles("testdata/findfiles", ".*.yml", "node_modules.*", nil)
		assert.NoError(err)
		assert.ElementsMatch([]string{
			"testdata/findfiles/a/some-file.yml",
			"testdata/findfiles/a/b/user.role.bogus.yml",
			"testdata/findfiles/user.admin.yml",
			"testdata/findfiles/user.role.admin.yml",
			"testdata/findfiles/user.role.author.yml",
			"testdata/findfiles/user.role.editor.yml",
		}, files)
	})

	t.Run("skipDir", func(t *testing.T) {
		files, err := FindFiles("testdata/findfiles", ".*.yml", "", []string{
			"node_modules", "a"})
		assert.NoError(err)
		assert.ElementsMatch([]string{
			"testdata/findfiles/user.admin.yml",
			"testdata/findfiles/user.role.admin.yml",
			"testdata/findfiles/user.role.author.yml",
			"testdata/findfiles/user.role.editor.yml",
		}, files)
	})

	t.Run("skipDir1Deeper", func(t *testing.T) {
		files, err := FindFiles("testdata/findfiles", ".*.yml", "", []string{
			"node_modules", "a/b"})
		assert.NoError(err)
		assert.ElementsMatch([]string{
			"testdata/findfiles/a/some-file.yml",
			"testdata/findfiles/user.admin.yml",
			"testdata/findfiles/user.role.admin.yml",
			"testdata/findfiles/user.role.author.yml",
			"testdata/findfiles/user.role.editor.yml",
		}, files)
	})
}

func TestMergeBoolPtrs(t *testing.T) {
	assert := assert.New(t)

	bTrue := true
	bFalse := false
	var boolVarA *bool
	var boolVarB *bool

	assert.Nil(boolVarA)
	assert.Nil(boolVarB)

	boolVarA = &bTrue
	MergeBoolPtrs(boolVarA, boolVarB)
	assert.True(*boolVarA)

	boolVarB = &bFalse
	MergeBoolPtrs(boolVarA, boolVarB)
	assert.False(*boolVarA)
}

func TestMergeString(t *testing.T) {
	assert := assert.New(t)

	strVarA := "foo"
	strVarB := ""
	MergeString(&strVarA, strVarB)
	assert.Equal("foo", strVarA)

	strVarB = "bar"
	MergeString(&strVarA, strVarB)
	assert.Equal("bar", strVarA)
}

func TestMergeStringSlice(t *testing.T) {
	assert := assert.New(t)

	strSlcVarA := []string{"foo", "bar"}
	strSlcVarB := []string(nil)

	MergeStringSlice(&strSlcVarA, strSlcVarB)
	assert.EqualValues([]string{"foo", "bar"}, strSlcVarA)

	strSlcVarB = []string{"foo", "baz"}
	MergeStringSlice(&strSlcVarA, strSlcVarB)
	assert.EqualValues([]string{"foo", "baz"}, strSlcVarA)

	strSlcVarA = []string(nil)
	strSlcVarB = []string{"zoom", "zap"}
	MergeStringSlice(&strSlcVarA, strSlcVarB)
	assert.EqualValues([]string{"zoom", "zap"}, strSlcVarA)

}

func TestMergeIntSlice(t *testing.T) {
	assert := assert.New(t)

	slcVarA := []int{1, 2}
	SlcVarB := []int(nil)

	MergeIntSlice(&slcVarA, SlcVarB)
	assert.EqualValues([]int{1, 2}, slcVarA)

	SlcVarB = []int{1, 3}
	MergeIntSlice(&slcVarA, SlcVarB)
	assert.EqualValues([]int{1, 3}, slcVarA)

	slcVarA = []int(nil)
	SlcVarB = []int{4, 5}
	MergeIntSlice(&slcVarA, SlcVarB)
	assert.EqualValues([]int{4, 5}, slcVarA)

}

func TestStringSliceContains(t *testing.T) {
	assert := assert.New(t)
	assert.False(StringSliceContains([]string{}, "foo"))
	assert.False(StringSliceContains([]string{"bar"}, "foo"))
	assert.True(StringSliceContains([]string{"bar", "foo"}, "foo"))
}

func TestIntSliceContains(t *testing.T) {
	assert := assert.New(t)
	assert.False(IntSliceContains([]int{}, 10))
	assert.False(IntSliceContains([]int{5}, 10))
	assert.True(IntSliceContains([]int{5, 10}, 10))
}

func TestStringSlicesIntersect(t *testing.T) {
	intersect := StringSlicesIntersect(
		[]string{"foo"}, []string{})
	if len(intersect) != 0 {
		t.Errorf("Intersect should be empty, got '%+v'", intersect)
	}

	intersect = StringSlicesIntersect(
		[]string{"foo"}, []string{"bar"})
	if len(intersect) != 0 {
		t.Errorf("Intersect should be empty, got '%+v'", intersect)
	}

	intersect = StringSlicesIntersect(
		[]string{"foo"}, []string{"bar", "foo"})
	expectedIntersect := []string{"foo"}
	if len(intersect) != 1 || !reflect.DeepEqual(intersect, expectedIntersect) {
		t.Errorf("Intersect should have 1 item, got '%+v'", intersect)
	}

	intersect = StringSlicesIntersect(
		[]string{"foo", "baz", "zoom"}, []string{"bar", "foo", "zoo", "zoom"})
	expectedIntersect = []string{"foo", "zoom"}
	if len(intersect) != 2 || !reflect.DeepEqual(intersect, expectedIntersect) {
		t.Errorf("Intersect should have 2 item, got '%+v'", intersect)
	}
}

func TestStringIsUrl(t *testing.T) {
	isUrl := StringIsUrl("foo/bar.yml")
	if isUrl {
		t.Error("expected isUrl to be false, got true")
	}

	isUrl = StringIsUrl("~/foo/bar.yml")
	if isUrl {
		t.Error("expected isUrl to be false, got true")
	}

	isUrl = StringIsUrl("/home/user/foo/bar.yml")
	if isUrl {
		t.Error("expected isUrl to be false, got true")
	}

	isUrl = StringIsUrl("https://example.com/foo.yml")
	if !isUrl {
		t.Error("expected isUrl to be true, got false")
	}

	isUrl = StringIsUrl("https://127.0.0.1:8080/foo.yml")
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

	c, err := FetchContentFromUrl(svr.URL + "/foo.yml")
	if err != nil {
		t.Errorf("expected err to be nil got %v", err)
	}
	if string(c) != expected {
		t.Errorf("expected content to be %s, got %s", expected, c)
	}
}

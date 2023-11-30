package utils_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"testing"

	. "github.com/salsadigitalauorg/shipshape/pkg/utils"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
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

func TestLookupYamlPath(t *testing.T) {
	assert := assert.New(t)

	t.Run("newPathErr", func(t *testing.T) {
		n := yaml.Node{}
		yamlData := []byte(`
foo:
  bar: baz
`)
		yaml.Unmarshal(yamlData, &n)
		_, err := LookupYamlPath(&n, ")")
		assert.Error(err, "syntax error at position 0, following \"\"")
	})

	t.Run("valid", func(t *testing.T) {
		n := yaml.Node{}
		yamlData := []byte(`
foo:
  bar: baz
`)
		yaml.Unmarshal(yamlData, &n)
		ns, err := LookupYamlPath(&n, "foo.bar")
		assert.NoError(err)
		assert.Equal("baz", ns[0].Value)
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

func TestSliceContains(t *testing.T) {
	assertion := assert.New(t)
	assertion.False(SliceContains([]any{}, "foo"))
	assertion.False(SliceContains([]any{"bar"}, "foo"))
	assertion.True(SliceContains([]any{"bar", "foo"}, "foo"))
	assertion.False(SliceContains([]any{"true"}, true))
	assertion.False(SliceContains([]any{"true", "false"}, false))
	assertion.False(SliceContains([]any{"true", true}, false))
	assertion.False(SliceContains([]any{"true", true, 0, 1, nil, "false"}, false))
	assertion.True(SliceContains([]any{"true", true, 0, 1, nil, "false"}, nil))
	assertion.True(SliceContains([]any{"true", true, 0, 1, nil, "false"}, 0))
	assertion.True(SliceContains([]any{"true", true, 0, 1, nil, "false"}, 1))
	assertion.False(SliceContains([]any{"true", true, 0, 1, nil, "false"}, "1"))
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

func TestStringSlicesIntersectUnique(t *testing.T) {
	intersect := StringSlicesIntersectUnique(
		[]string{"foo"}, []string{})
	if len(intersect) != 0 {
		t.Errorf("Intersect should be empty, got '%+v'", intersect)
	}

	intersect = StringSlicesIntersectUnique(
		[]string{"foo"}, []string{"bar"})
	if len(intersect) != 0 {
		t.Errorf("Intersect should be empty, got '%+v'", intersect)
	}

	intersect = StringSlicesIntersectUnique(
		[]string{"foo"}, []string{"bar", "foo"})
	expectedIntersect := []string{"foo"}
	if len(intersect) != 1 || !reflect.DeepEqual(intersect, expectedIntersect) {
		t.Errorf("Intersect should have 1 item, got '%+v'", intersect)
	}

	intersect = StringSlicesIntersectUnique(
		[]string{"foo", "baz", "zoom"}, []string{"bar", "foo", "zoo", "zoom"})
	sort.Strings(intersect)
	expectedIntersect = []string{"foo", "zoom"}
	if len(intersect) != 2 || !reflect.DeepEqual(intersect, expectedIntersect) {
		t.Errorf("Intersect should have 2 item, got '%+v'", intersect)
	}

	intersect = StringSlicesIntersectUnique(
		[]string{"foo", "baz", "zoom"}, []string{"bar", "foo", "zoo", "zoom", "foo", "bar", "zoom", "zoo"})
	sort.Strings(intersect)
	expectedIntersect = []string{"foo", "zoom"}
	if len(intersect) != 2 || !reflect.DeepEqual(intersect, expectedIntersect) {
		t.Errorf("Intersect should have 2 item, got '%+v'", intersect)
	}
}

func TestStringSlicesInterdiff(t *testing.T) {
	interdiff := StringSlicesInterdiff(
		[]string{"foo"}, []string{})
	if len(interdiff) != 0 {
		t.Errorf("Interdiff should be empty, got '%+v'", interdiff)
	}

	interdiff = StringSlicesInterdiff(
		[]string{"foo"}, []string{"foo"})
	if len(interdiff) != 0 {
		t.Errorf("Interdiff should be empty, got '%+v'", interdiff)
	}

	interdiff = StringSlicesInterdiff(
		[]string{"foo"}, []string{"bar", "foo"})
	expectedInterdiff := []string{"bar"}
	if len(interdiff) != 1 || !reflect.DeepEqual(interdiff, expectedInterdiff) {
		t.Errorf("Interdiff should have 1 item, got '%+v'", interdiff)
	}

	interdiff = StringSlicesInterdiff(
		[]string{"foo", "baz", "zoom"}, []string{"bar", "foo", "zoo", "zoom", "bar"})
	sort.Strings(interdiff)
	expectedInterdiff = []string{"bar", "bar", "zoo"}
	if len(interdiff) != 3 || !reflect.DeepEqual(interdiff, expectedInterdiff) {
		t.Errorf("Interdiff should have 4 item, got '%+v'", interdiff)
	}
}

func TestStringSlicesInterdiffUnique(t *testing.T) {
	interdiff := StringSlicesInterdiffUnique(
		[]string{"foo"}, []string{})
	if len(interdiff) != 0 {
		t.Errorf("Interdiff should be empty, got '%+v'", interdiff)
	}

	interdiff = StringSlicesInterdiffUnique(
		[]string{"foo"}, []string{"foo"})
	if len(interdiff) != 0 {
		t.Errorf("Interdiff should be empty, got '%+v'", interdiff)
	}

	interdiff = StringSlicesInterdiffUnique(
		[]string{"foo"}, []string{"bar", "foo"})
	expectedInterdiff := []string{"bar"}
	if len(interdiff) != 1 || !reflect.DeepEqual(interdiff, expectedInterdiff) {
		t.Errorf("Interdiff should have 1 item, got '%+v'", interdiff)
	}

	interdiff = StringSlicesInterdiffUnique(
		[]string{"foo", "baz", "zoom"}, []string{"bar", "foo", "zoo", "zoom", "bar"})
	sort.Strings(interdiff)
	expectedInterdiff = []string{"bar", "zoo"}
	if len(interdiff) != 2 || !reflect.DeepEqual(interdiff, expectedInterdiff) {
		t.Errorf("Interdiff should have 2 item, got '%+v'", interdiff)
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

func TestIsDirectory(t *testing.T) {
	if a, e := IsDirectory("testdata"); e != nil || !a {
		t.Errorf("expected directory 'testdata' to exist")
	}
	if a, _ := IsDirectory("nodir"); a {
		t.Errorf("expected directory 'nodir' to not exist")
	}
}

func TestFileContains(t *testing.T) {
	if a, _ := FileContains("testdata/filecontains/index.php", "use Drupal\\Core\\DrupalKernel;"); !a {
		t.Errorf("expected 'DrupalKernel' in index.php")
	}
	if b, _ := FileContains("testdata/filecontains/index.php", "notfound"); b {
		t.Errorf("expected 'notfound' to not appear in index.php")
	}
}

func TestHasComposerDependency(t *testing.T) {
	deps := []string{"laravel/framework"}
	if a, _ := HasComposerDependency("testdata/composer", deps); !a {
		t.Errorf("expected 'laravel/framework' to be found in composer.json")
	}
	deps = []string{"notfound/notfound"}
	if a, _ := HasComposerDependency("testdata/composer", deps); a {
		t.Errorf("expected 'notfound/notfound' to not be found in composer.json")
	}
	if _, err := HasComposerDependency("testdata/filecontains", deps); err == nil {
		t.Errorf("expected file not found got %s", err)
	}
}

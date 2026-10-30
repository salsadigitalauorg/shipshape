package main

import (
	"bufio"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// checksumsFixture mirrors the real GoReleaser output format: lowercase hex
// digest, two spaces, artifact name. Both the bare binaries and the .tar.gz
// archives are present, as they are on every published release.
const checksumsFixture = `1f91083964b8d9d21742efaf25f99f780f90ef51b60cf8d9f82e4bd57847aa96  shipshape-Darwin-aarch64
5e50e0d0706b1b0331bdcd97fd2cb5cc4b16561dbb69761cfcdc2a2482721458  shipshape-Darwin-aarch64.tar.gz
0eabe29a91ba6d3ad77f7681a8a124967d841ce79c83541e8387ab7044096c97  shipshape-Darwin-x86_64
eef4d7f0e0ce63a57319e044215edc322581b1c66a8a6d56ebbb52db34b2ea47  shipshape-Darwin-x86_64.tar.gz
3daea2c60f27f190b336dfffb383f3934b04402ad4e1b51ad260949de887b3dd  shipshape-Linux-aarch64
4ea1590df43c41fc3b23333c67281e400cff6b45e3f3be2da2cf79de1517bf2b  shipshape-Linux-aarch64.tar.gz
6a4935fddf7f5c208f40c652115262e9fe7fdf69fa5f25c299bddcd4c246b85e  shipshape-Linux-x86_64
c607daa08b8ec33aef08c820bad2504014f39cf6e962c6685b87885c11e79532  shipshape-Linux-x86_64.tar.gz
`

func parse(t *testing.T, s string) map[string]string {
	t.Helper()
	sums, err := ParseChecksums(bufio.NewScanner(strings.NewReader(s)))
	require.NoError(t, err)
	return sums
}

// TestClassS pins the port of Homebrew's Formulary.class_s. The expected
// values were derived by running the real Ruby implementation from
// Library/Homebrew/formulary.rb against each input. A mismatch between a
// formula's class name and its file name makes the formula unloadable, so
// these cases must not be loosened to match the implementation.
func TestClassS(t *testing.T) {
	tt := []struct {
		name     string
		in       string
		expected string
	}{
		{"plain", "shipshape", "Shipshape"},
		{"versionedAlpha", "shipshape@1.0.0-alpha.1.6", "ShipshapeAT100Alpha16"},
		{"versionedAlphaPatch", "shipshape@1.0.0-alpha.1.5.2", "ShipshapeAT100Alpha152"},
		{"majorMinor", "shipshape@1.0", "ShipshapeAT10"},
		{"hyphenated", "foo-bar", "FooBar"},
		{"underscored", "foo_bar", "FooBar"},
		{"dotted", "foo.bar", "FooBar"},
		{"plus", "foo+", "Foox"},
		{"mixedCaseIsDowncased", "ShipShape", "Shipshape"},
		{"empty", "", ""},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ClassS(tc.in))
		})
	}
}

// TestClassSMatchesRuby runs the real Homebrew implementation and compares it
// against the Go port. It is skipped when Ruby is unavailable (CI on
// ubuntu-latest has it; a minimal container may not), so TestClassS above
// remains the authoritative pinned-value check.
func TestClassSMatchesRuby(t *testing.T) {
	ruby, err := exec.LookPath("ruby")
	if err != nil {
		t.Skip("ruby not available; skipping differential check")
	}

	// Verbatim from Library/Homebrew/formulary.rb.
	const script = `
def class_s(name)
  class_name = name.capitalize
  class_name.gsub!(/[-_.\s][a-zA-Z0-9]/) { |m| m.chars.fetch(-1).upcase }
  class_name.tr!("+", "x")
  class_name.sub!(/(.)@(\d)/, "\\1AT\\2")
  class_name
end
ARGV.each { |n| puts class_s(n) }
`

	names := []string{
		"shipshape",
		"shipshape@1.0.0-alpha.1.6",
		"shipshape@1.0.0-alpha.1.5.2",
		"shipshape@1.0",
		"shipshape@2.0.0",
		"foo-bar",
		"foo_bar",
		"foo.bar",
		"foo+",
		"ShipShape",
	}

	args := append([]string{"-e", script}, names...)
	out, err := exec.Command(ruby, args...).Output()
	require.NoError(t, err)

	want := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
	require.Len(t, want, len(names))

	for i, name := range names {
		assert.Equal(t, want[i], ClassS(name),
			"ClassS(%q) must match Homebrew's Formulary.class_s", name)
	}
}

func TestParseChecksums(t *testing.T) {
	t.Run("validFixture", func(t *testing.T) {
		sums := parse(t, checksumsFixture)
		assert.Len(t, sums, 8)
		assert.Equal(t,
			"5e50e0d0706b1b0331bdcd97fd2cb5cc4b16561dbb69761cfcdc2a2482721458",
			sums["shipshape-Darwin-aarch64.tar.gz"])
	})

	t.Run("blankLinesIgnored", func(t *testing.T) {
		sums := parse(t, "\n"+checksumsFixture+"\n\n")
		assert.Len(t, sums, 8)
	})

	t.Run("digestIsLowercased", func(t *testing.T) {
		sums := parse(t,
			"AAAABBBBCCCCDDDDEEEEFFFF00001111222233334444555566667777888899990  x\n"[:64]+
				"  shipshape-Linux-x86_64.tar.gz\n")
		assert.Equal(t,
			strings.ToLower("AAAABBBBCCCCDDDDEEEEFFFF0000111122223333444455556666777788889999"),
			sums["shipshape-Linux-x86_64.tar.gz"])
	})

	t.Run("malformedLineErrors", func(t *testing.T) {
		_, err := ParseChecksums(bufio.NewScanner(strings.NewReader("not a checksum line\n")))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "malformed")
	})

	t.Run("shortDigestErrors", func(t *testing.T) {
		_, err := ParseChecksums(bufio.NewScanner(strings.NewReader("abc123  foo.tar.gz\n")))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "malformed")
	})

	t.Run("emptyFileErrors", func(t *testing.T) {
		_, err := ParseChecksums(bufio.NewScanner(strings.NewReader("")))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no entries")
	})
}

func baseOptions() Options {
	return Options{
		Tag:     "v1.0.0-alpha.1.6",
		Repo:    "salsadigitalauorg/shipshape",
		Name:    "shipshape",
		Branch:  "1.x",
		Desc:    "Golang CLI tool to run checks/audits on your systems",
		License: "MIT",
	}
}

func TestRenderRequiredOptions(t *testing.T) {
	sums := parse(t, checksumsFixture)

	tt := []struct {
		name    string
		mutate  func(*Options)
		wantErr string
	}{
		{"noTag", func(o *Options) { o.Tag = "" }, "tag is required"},
		{"noRepo", func(o *Options) { o.Repo = "" }, "repo is required"},
		{"noName", func(o *Options) { o.Name = "" }, "name is required"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			opts := baseOptions()
			tc.mutate(&opts)
			_, err := Render(opts, sums)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

// TestRenderMissingArchiveFails is the important failure mode: a formula with
// an empty sha256 installs nothing and fails opaquely for users, so a missing
// archive must break the release rather than publish a broken formula.
func TestRenderMissingArchiveFails(t *testing.T) {
	sums := parse(t, checksumsFixture)
	delete(sums, "shipshape-Linux-aarch64.tar.gz")

	_, err := Render(baseOptions(), sums)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required archives")
	assert.Contains(t, err.Error(), "shipshape-Linux-aarch64.tar.gz")
}

func TestRenderFilenames(t *testing.T) {
	files, err := Render(baseOptions(), parse(t, checksumsFixture))
	require.NoError(t, err)
	require.Len(t, files, 2)

	assert.Equal(t, "shipshape.rb", files[0].Filename)
	assert.Equal(t, "shipshape@1.0.0-alpha.1.6.rb", files[1].Filename)
}

func TestRenderMainFormula(t *testing.T) {
	files, err := Render(baseOptions(), parse(t, checksumsFixture))
	require.NoError(t, err)
	main := string(files[0].Content)

	t.Run("classMatchesFilename", func(t *testing.T) {
		assert.Contains(t, main, "class Shipshape < Formula")
	})

	t.Run("versionHasNoVPrefix", func(t *testing.T) {
		assert.Contains(t, main, `version "1.0.0-alpha.1.6"`)
		assert.NotContains(t, main, `version "v1.0.0-alpha.1.6"`)
	})

	t.Run("hasHeadBlockForSourceBuild", func(t *testing.T) {
		assert.Contains(t, main, "head do")
		assert.Contains(t, main,
			`url "https://github.com/salsadigitalauorg/shipshape.git", branch: "1.x"`)
		assert.Contains(t, main, `depends_on "go" => :build`)
	})

	// `brew audit` fails with "`head` should be put before `on_macos`" if
	// these are ordered the other way round.
	t.Run("headPrecedesPlatformBlocks", func(t *testing.T) {
		head := strings.Index(main, "head do")
		macos := strings.Index(main, "on_macos do")
		require.Positive(t, head)
		require.Positive(t, macos)
		assert.Less(t, head, macos,
			"brew audit requires head before on_macos")
	})

	t.Run("sourceBuildGeneratesBeforeCompiling", func(t *testing.T) {
		// registry_gen.go is gitignored; a plain go build on a fresh
		// checkout fails without this.
		assert.Contains(t, main, `system "go", "generate", "./..."`)
		gen := strings.Index(main, `"generate"`)
		build := strings.Index(main, `system "go", "build"`)
		require.Positive(t, gen)
		require.Positive(t, build)
		assert.Less(t, gen, build, "go generate must precede go build")
	})

	t.Run("injectsVersionLdflags", func(t *testing.T) {
		assert.Contains(t, main,
			"-X github.com/salsadigitalauorg/shipshape/cmd.version=")
		assert.Contains(t, main,
			"-X github.com/salsadigitalauorg/shipshape/cmd.commit=")
	})

	t.Run("isNotKegOnly", func(t *testing.T) {
		assert.NotContains(t, main, "keg_only")
	})

	t.Run("allFourPlatformsPresent", func(t *testing.T) {
		for _, want := range []string{
			"shipshape-Darwin-x86_64.tar.gz",
			"shipshape-Darwin-aarch64.tar.gz",
			"shipshape-Linux-x86_64.tar.gz",
			"shipshape-Linux-aarch64.tar.gz",
		} {
			assert.Contains(t, main, want)
		}
		assert.Equal(t, 4, strings.Count(main, "sha256 "))
	})

	t.Run("urlsPointAtTagNotVersion", func(t *testing.T) {
		assert.Contains(t, main,
			"https://github.com/salsadigitalauorg/shipshape/releases/download/v1.0.0-alpha.1.6/")
	})

	t.Run("platformBlocksAreCorrectlyNested", func(t *testing.T) {
		macos := strings.Index(main, "on_macos do")
		linux := strings.Index(main, "on_linux do")
		require.Positive(t, macos)
		require.Positive(t, linux)
		assert.Less(t, macos, linux)

		// The Darwin archives must appear between on_macos and on_linux.
		macosBlock := main[macos:linux]
		assert.Contains(t, macosBlock, "shipshape-Darwin-x86_64.tar.gz")
		assert.Contains(t, macosBlock, "shipshape-Darwin-aarch64.tar.gz")
		assert.NotContains(t, macosBlock, "shipshape-Linux")
	})
}

func TestRenderVersionedFormula(t *testing.T) {
	files, err := Render(baseOptions(), parse(t, checksumsFixture))
	require.NoError(t, err)
	pinned := string(files[1].Content)

	t.Run("classMatchesFilename", func(t *testing.T) {
		// Must agree with ClassS("shipshape@1.0.0-alpha.1.6").
		assert.Contains(t, pinned, "class ShipshapeAT100Alpha16 < Formula")
	})

	t.Run("isKegOnly", func(t *testing.T) {
		assert.Contains(t, pinned, "keg_only :versioned_formula")
	})

	t.Run("hasNoHeadBlock", func(t *testing.T) {
		assert.NotContains(t, pinned, "head do")
		assert.NotContains(t, pinned, "build.head?")
		assert.NotContains(t, pinned, `depends_on "go" => :build`)
	})

	t.Run("installsBinaryDirectly", func(t *testing.T) {
		assert.Contains(t, pinned, `bin.install "shipshape"`)
		assert.NotContains(t, pinned, `system "go", "build"`)
	})

	t.Run("pinsSameVersionAsMain", func(t *testing.T) {
		assert.Contains(t, pinned, `version "1.0.0-alpha.1.6"`)
	})
}

// TestRenderedClassNamesAgreeWithFilenames guards the invariant that actually
// breaks installs: Homebrew derives the expected class from the file name, so
// every rendered file's declared class must equal ClassS(basename).
func TestRenderedClassNamesAgreeWithFilenames(t *testing.T) {
	for _, tag := range []string{
		"v1.0.0-alpha.1.6",
		"v1.0.0-alpha.1.5.2",
		"v1.2.3",
		"v0.3.1",
	} {
		t.Run(tag, func(t *testing.T) {
			opts := baseOptions()
			opts.Tag = tag

			files, err := Render(opts, parse(t, checksumsFixture))
			require.NoError(t, err)

			for _, f := range files {
				expected := "class " + ClassS(strings.TrimSuffix(f.Filename, ".rb")) + " < Formula"
				assert.Contains(t, string(f.Content), expected,
					"%s must declare the class Homebrew derives from its name", f.Filename)
			}
		})
	}
}

func TestWrite(t *testing.T) {
	files, err := Render(baseOptions(), parse(t, checksumsFixture))
	require.NoError(t, err)

	t.Run("createsNestedDirAndFiles", func(t *testing.T) {
		dir := t.TempDir() + "/Formula"
		require.NoError(t, Write(dir, files))

		for _, f := range files {
			assert.FileExists(t, dir+"/"+f.Filename)
		}
	})

	t.Run("overwritesExistingFormula", func(t *testing.T) {
		// cmd/gen's createFile returns early when the target exists; a
		// formula must be replaced on every release instead.
		dir := t.TempDir()
		require.NoError(t, Write(dir, []Rendered{
			{Filename: "shipshape.rb", Content: []byte("stale")},
		}))
		require.NoError(t, Write(dir, files))

		raw, err := os.ReadFile(dir + "/shipshape.rb")
		require.NoError(t, err)
		assert.NotEqual(t, "stale", string(raw))
		assert.Contains(t, string(raw), "class Shipshape < Formula")
	})
}

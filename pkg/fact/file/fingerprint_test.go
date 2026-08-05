package file_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	. "github.com/salsadigitalauorg/shipshape/pkg/fact/file"
	"github.com/salsadigitalauorg/shipshape/pkg/internal"
	"github.com/stretchr/testify/assert"
)

func TestFileFingerprintInit(t *testing.T) {
	assert := assert.New(t)

	factPlugin := fact.Manager().GetFactories()["file:fingerprint"]("testFileFingerprint")
	assert.NotNil(factPlugin)
	fingerprintFacter, ok := factPlugin.(*Fingerprint)
	assert.True(ok)
	assert.Equal("testFileFingerprint", fingerprintFacter.GetId())
}

func TestFileFingerprintPluginName(t *testing.T) {
	fingerprint := NewFingerprint("testFileFingerprint")
	assert.Equal(t, "file:fingerprint", fingerprint.GetName())
}

func TestFileFingerprintCollect(t *testing.T) {
	tests := []internal.FactCollectTest{
		{
			Name: "testNoFrameworks",
			FactFn: func() fact.Facter {
				f := NewFingerprint("testNoFrameworks")
				f.Path = "testdata/fingerprint/markers-only"
				return f
			},
			ExpectedFormat: "",
			ExpectedErrors: []error{ErrNoFrameworks},
		},
		{
			Name: "testNoPath",
			FactFn: func() fact.Facter {
				f := NewFingerprint("testNoPath")
				f.Frameworks = map[string]FrameworkSignature{
					"drupal": {Markers: []string{"use Drupal\\Core\\DrupalKernel;"}},
				}
				return f
			},
			ExpectedFormat: "",
			ExpectedErrors: []error{ErrNoPath},
		},
		{
			Name: "testMarkersOnlyAboveThreshold",
			FactFn: func() fact.Facter {
				f := NewFingerprint("testMarkersOnlyAboveThreshold")
				f.Path = "testdata/fingerprint/markers-only"
				f.Threshold = 1
				f.Frameworks = map[string]FrameworkSignature{
					"drupal": {
						Markers: []string{"use Drupal\\Core\\DrupalKernel;"},
					},
				}
				return f
			},
			ExpectedFormat: data.FormatMapString,
			ExpectedData: map[string]string{
				"drupal": "5",
			},
			ExpectedErrors: []error{},
		},
		{
			Name: "testMarkersOnlyBelowThreshold",
			FactFn: func() fact.Facter {
				f := NewFingerprint("testMarkersOnlyBelowThreshold")
				f.Path = "testdata/fingerprint/markers-only"
				f.Threshold = 10
				f.Frameworks = map[string]FrameworkSignature{
					"drupal": {
						Markers: []string{"use Drupal\\Core\\DrupalKernel;"},
					},
				}
				return f
			},
			ExpectedFormat: data.FormatMapString,
			ExpectedData:   map[string]string{},
			ExpectedErrors: []error{},
		},
		{
			Name: "testDirsOnly",
			FactFn: func() fact.Facter {
				f := NewFingerprint("testDirsOnly")
				f.Path = "testdata/fingerprint/dirs-only"
				f.Threshold = 1
				f.Frameworks = map[string]FrameworkSignature{
					"drupal": {
						Dirs: []string{"web", "docroot"},
					},
				}
				return f
			},
			ExpectedFormat: data.FormatMapString,
			ExpectedData: map[string]string{
				"drupal": "5",
			},
			ExpectedErrors: []error{},
		},
		{
			Name: "testMarkersAndDirsAccumulate",
			FactFn: func() fact.Facter {
				f := NewFingerprint("testMarkersAndDirsAccumulate")
				f.Path = "testdata/fingerprint/combined"
				f.Threshold = 1
				f.Entrypoints = []string{"index.php"}
				f.Frameworks = map[string]FrameworkSignature{
					"drupal": {
						Markers: []string{"use Drupal\\Core\\DrupalKernel;"},
						Dirs:    []string{"web", "docroot"},
					},
				}
				return f
			},
			ExpectedFormat: data.FormatMapString,
			ExpectedData: map[string]string{
				// markers: 1 marker * 1 entrypoint * weight 5 = 5
				// dirs: 1 matching dir * weight 5 = 5
				"drupal": "10",
			},
			ExpectedErrors: []error{},
		},
		{
			Name: "testThresholdBoundaryExactlyAtThresholdDoesNotBreach",
			FactFn: func() fact.Facter {
				f := NewFingerprint("testThresholdBoundaryExactlyAtThresholdDoesNotBreach")
				f.Path = "testdata/fingerprint/combined"
				f.Threshold = 10
				f.Entrypoints = []string{"index.php"}
				f.Frameworks = map[string]FrameworkSignature{
					"drupal": {
						Markers: []string{"use Drupal\\Core\\DrupalKernel;"},
						Dirs:    []string{"web", "docroot"},
					},
				}
				return f
			},
			ExpectedFormat: data.FormatMapString,
			ExpectedData:   map[string]string{},
			ExpectedErrors: []error{},
		},
		{
			Name: "testThresholdBoundaryJustOverThresholdBreaches",
			FactFn: func() fact.Facter {
				f := NewFingerprint("testThresholdBoundaryJustOverThresholdBreaches")
				f.Path = "testdata/fingerprint/combined"
				f.Threshold = 9
				f.Entrypoints = []string{"index.php"}
				f.Frameworks = map[string]FrameworkSignature{
					"drupal": {
						Markers: []string{"use Drupal\\Core\\DrupalKernel;"},
						Dirs:    []string{"web", "docroot"},
					},
				}
				return f
			},
			ExpectedFormat: data.FormatMapString,
			ExpectedData: map[string]string{
				"drupal": "10",
			},
			ExpectedErrors: []error{},
		},
		{
			Name: "testCustomWeights",
			FactFn: func() fact.Facter {
				f := NewFingerprint("testCustomWeights")
				f.Path = "testdata/fingerprint/combined"
				f.Threshold = 1
				f.Entrypoints = []string{"index.php"}
				f.Weights = FingerprintWeights{Markers: 100, Dirs: 1}
				f.Frameworks = map[string]FrameworkSignature{
					"drupal": {
						Markers: []string{"use Drupal\\Core\\DrupalKernel;"},
						Dirs:    []string{"web", "docroot"},
					},
				}
				return f
			},
			ExpectedFormat: data.FormatMapString,
			ExpectedData: map[string]string{
				// markers: 1 * 1 * 100 = 100; dirs: 1 * 1 = 1; total 101
				"drupal": "101",
			},
			ExpectedErrors: []error{},
		},
		{
			Name: "testMultiFramework",
			FactFn: func() fact.Facter {
				f := NewFingerprint("testMultiFramework")
				f.Path = "testdata/fingerprint/multi-framework"
				f.Threshold = 1
				f.Entrypoints = []string{"index.php"}
				f.Frameworks = map[string]FrameworkSignature{
					"drupal": {
						Markers: []string{"use Drupal\\Core\\DrupalKernel;"},
						Dirs:    []string{"web", "docroot"},
					},
					"wordpress": {
						Markers: []string{"require_once __DIR__ . '/wp-load.php';"},
						Dirs:    []string{"wp-content"},
					},
					"symfony": {
						Markers: []string{"this marker never matches"},
						Dirs:    []string{"this-dir-never-matches"},
					},
				}
				return f
			},
			ExpectedFormat: data.FormatMapString,
			ExpectedData: map[string]string{
				"drupal":    "10",
				"wordpress": "10",
			},
			ExpectedErrors: []error{},
		},
		{
			Name: "testMarkerContentNotLeakedIntoOutput",
			FactFn: func() fact.Facter {
				f := NewFingerprint("testMarkerContentNotLeakedIntoOutput")
				f.Path = "testdata/fingerprint/markers-only"
				f.Threshold = 1
				f.Frameworks = map[string]FrameworkSignature{
					"drupal": {
						Markers: []string{"use Drupal\\Core\\DrupalKernel;"},
					},
				}
				return f
			},
			ExpectedFormat: data.FormatMapString,
			ExpectedData: map[string]string{
				"drupal": "5",
			},
			ExpectedErrors: []error{},
		},
		{
			Name: "testDependenciesOnlyRequire",
			FactFn: func() fact.Facter {
				f := NewFingerprint("testDependenciesOnlyRequire")
				f.Path = "testdata/fingerprint/dependencies-only"
				f.Threshold = 1
				f.Frameworks = map[string]FrameworkSignature{
					"drupal": {
						Dependencies: []string{"drupal/core-recommended"},
					},
				}
				return f
			},
			ExpectedFormat: data.FormatMapString,
			ExpectedData: map[string]string{
				"drupal": "10",
			},
			ExpectedErrors: []error{},
		},
		{
			Name: "testDependenciesRequireDevFixesZeroXGap",
			FactFn: func() fact.Facter {
				f := NewFingerprint("testDependenciesRequireDevFixesZeroXGap")
				f.Path = "testdata/fingerprint/dependencies-require-dev-only"
				f.Threshold = 1
				f.Frameworks = map[string]FrameworkSignature{
					"drupal": {
						Dependencies: []string{"drupal/core-dev"},
					},
				}
				return f
			},
			ExpectedFormat: data.FormatMapString,
			ExpectedData: map[string]string{
				"drupal": "10",
			},
			ExpectedErrors: []error{},
		},
		{
			Name: "testDependencyMatchDoesNotAccumulatePerHit",
			FactFn: func() fact.Facter {
				f := NewFingerprint("testDependencyMatchDoesNotAccumulatePerHit")
				f.Path = "testdata/fingerprint/dependencies-only"
				f.Threshold = 1
				f.Frameworks = map[string]FrameworkSignature{
					"drupal": {
						// Both configured dependency names match (php and
						// drupal/core-recommended are both present in the
						// fixture's require block), but the score must
						// still be a single weights.dependencies award,
						// not two.
						Dependencies: []string{"php", "drupal/core-recommended"},
					},
				}
				return f
			},
			ExpectedFormat: data.FormatMapString,
			ExpectedData: map[string]string{
				"drupal": "10",
			},
			ExpectedErrors: []error{},
		},
		{
			Name: "testAllThreeSignalsCombined",
			FactFn: func() fact.Facter {
				f := NewFingerprint("testAllThreeSignalsCombined")
				f.Path = "testdata/fingerprint/all-signals"
				f.Threshold = 1
				f.Entrypoints = []string{"index.php"}
				f.Frameworks = map[string]FrameworkSignature{
					"drupal": {
						Markers:      []string{"use Drupal\\Core\\DrupalKernel;"},
						Dirs:         []string{"web", "docroot"},
						Dependencies: []string{"drupal/core-recommended"},
					},
				}
				return f
			},
			ExpectedFormat: data.FormatMapString,
			ExpectedData: map[string]string{
				// markers: 1*1*5 = 5; dirs: 1*5 = 5; dependencies: 10;
				// total 20.
				"drupal": "20",
			},
			ExpectedErrors: []error{},
		},
		{
			Name: "testMissingManifestIsCollectionError",
			FactFn: func() fact.Facter {
				f := NewFingerprint("testMissingManifestIsCollectionError")
				f.Path = "testdata/fingerprint/markers-only"
				f.Threshold = 1
				f.Frameworks = map[string]FrameworkSignature{
					"drupal": {
						Dependencies: []string{"drupal/core-recommended"},
					},
				}
				return f
			},
			ExpectedFormat: "",
			ExpectedErrors: []error{ErrManifestNotFound},
		},
		{
			Name: "testMalformedManifestIsCollectionError",
			FactFn: func() fact.Facter {
				f := NewFingerprint("testMalformedManifestIsCollectionError")
				f.Path = "testdata/fingerprint/malformed-manifest"
				f.Threshold = 1
				f.Frameworks = map[string]FrameworkSignature{
					"drupal": {
						Dependencies: []string{"drupal/core-recommended"},
					},
				}
				return f
			},
			ExpectedFormat: "",
			ExpectedErrors: []error{ErrManifestInvalidJSON},
		},
		{
			Name: "testInvalidDependencyPathExpression",
			FactFn: func() fact.Facter {
				f := NewFingerprint("testInvalidDependencyPathExpression")
				f.Path = "testdata/fingerprint/dependencies-only"
				f.Threshold = 1
				f.DependencyPaths = []string{"$.require[?(("}
				f.Frameworks = map[string]FrameworkSignature{
					"drupal": {
						Dependencies: []string{"drupal/core-recommended"},
					},
				}
				return f
			},
			ExpectedFormat: "",
			ExpectedErrors: []error{ErrInvalidDependencyPath},
		},
		{
			Name: "testCustomManifestAndDependencyPathsForPackageJson",
			FactFn: func() fact.Facter {
				f := NewFingerprint("testCustomManifestAndDependencyPathsForPackageJson")
				f.Path = "testdata/fingerprint/package-json-deps"
				f.Threshold = 1
				f.Manifest = "package.json"
				f.DependencyPaths = []string{"$.dependencies", "$.devDependencies"}
				f.Frameworks = map[string]FrameworkSignature{
					"react-app": {
						Dependencies: []string{"react"},
					},
					"jest-tested": {
						Dependencies: []string{"jest"},
					},
				}
				return f
			},
			ExpectedFormat: data.FormatMapString,
			ExpectedData: map[string]string{
				"react-app":   "10",
				"jest-tested": "10",
			},
			ExpectedErrors: []error{},
		},
		{
			Name: "testDependenciesSkippedWhenNoFrameworkUsesThem",
			FactFn: func() fact.Facter {
				// No composer.json under this path - if the dependencies
				// signal were unconditionally evaluated, this would fail
				// with ErrManifestNotFound even though no framework
				// configures 'dependencies'.
				f := NewFingerprint("testDependenciesSkippedWhenNoFrameworkUsesThem")
				f.Path = "testdata/fingerprint/markers-only"
				f.Threshold = 1
				f.Frameworks = map[string]FrameworkSignature{
					"drupal": {
						Markers: []string{"use Drupal\\Core\\DrupalKernel;"},
					},
				}
				return f
			},
			ExpectedFormat: data.FormatMapString,
			ExpectedData: map[string]string{
				"drupal": "5",
			},
			ExpectedErrors: []error{},
		},
	}

	config.ProjectDir = ""
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			internal.TestFactCollect(t, tt)
		})
	}
}

// TestFileFingerprintCollectDeterminism asserts that the emitted map is
// built identically across repeated runs when multiple frameworks are
// configured, since Go randomises map iteration and the fact must sort
// framework names during scoring to avoid flaky behaviour (e.g. around
// logging or any future short-circuiting).
func TestFileFingerprintCollectDeterminism(t *testing.T) {
	assert := assert.New(t)
	config.ProjectDir = ""

	var first map[string]string
	for i := 0; i < 25; i++ {
		f := NewFingerprint("testDeterminism")
		f.Path = "testdata/fingerprint/multi-framework"
		f.Threshold = 1
		f.Entrypoints = []string{"index.php"}
		f.Frameworks = map[string]FrameworkSignature{
			"drupal": {
				Markers: []string{"use Drupal\\Core\\DrupalKernel;"},
				Dirs:    []string{"web", "docroot"},
			},
			"wordpress": {
				Markers: []string{"require_once __DIR__ . '/wp-load.php';"},
				Dirs:    []string{"wp-content"},
			},
		}

		f.Collect()
		assert.Empty(f.GetErrors())

		got, ok := f.GetData().(map[string]string)
		assert.True(ok)

		if first == nil {
			first = got
			continue
		}
		assert.Equal(first, got, "fingerprint output changed between runs")
	}

	assert.Equal(map[string]string{
		"drupal":    "10",
		"wordpress": "10",
	}, first)
}

// TestFileFingerprintUnreadableEntrypoint asserts that an entrypoint which
// cannot be read is reported as a collection error rather than silently
// skipped. Skipping it would lose any marker it might have matched and
// under-score the framework, reporting a framework that is present as
// absent - the worst failure mode for an audit tool, and a regression this
// test locks down.
//
// The fixture is created at runtime with mode 0000 rather than committed,
// since a permission-stripped file in testdata would not survive checkout
// reliably.
func TestFileFingerprintUnreadableEntrypoint(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root: mode 0000 is still readable, cannot test unreadable file")
	}

	assert := assert.New(t)

	dir := t.TempDir()
	entrypoint := filepath.Join(dir, "index.php")
	marker := "use Drupal\\Core\\DrupalKernel;"
	if err := os.WriteFile(entrypoint, []byte("<?php\n"+marker+"\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	if err := os.Chmod(entrypoint, 0o000); err != nil {
		t.Fatalf("chmod fixture: %v", err)
	}
	t.Cleanup(func() {
		// Restore permissions so t.TempDir cleanup can remove the file.
		_ = os.Chmod(entrypoint, 0o644)
	})

	config.ProjectDir = ""

	f := NewFingerprint("testUnreadableEntrypoint")
	f.Path = dir
	// Threshold 4 against a marker weight of 5: were the marker matched,
	// this would score 5 and be emitted. The assertion below therefore
	// distinguishes "error raised" from "silently scored zero".
	f.Threshold = 4
	f.Entrypoints = []string{"index.php"}
	f.Frameworks = map[string]FrameworkSignature{
		"drupal": {Markers: []string{marker}},
	}

	f.Collect()

	assert.NotEmpty(f.GetErrors(), "unreadable entrypoint must raise a collection error")
	assert.Empty(f.GetData(), "no data should be emitted when collection fails")
}

// TestFileFingerprintMarkerMatchingIgnoresCarriageReturns asserts that a
// marker matches whole lines irrespective of LF vs CRLF line endings.
// Markers are matched against lines split on \n, so without trimming the
// trailing \r a config would silently stop matching against a file
// committed with Windows line endings - and 0.x's bufio.Scanner-based
// utils.FileContains stripped them, so not trimming would also be a
// behavioural regression against the check this replaces.
func TestFileFingerprintMarkerMatchingIgnoresCarriageReturns(t *testing.T) {
	assert := assert.New(t)

	marker := "use Drupal\\Core\\DrupalKernel;"

	for _, tc := range []struct {
		name    string
		content string
	}{
		{name: "lf", content: "<?php\n" + marker + "\n"},
		{name: "crlf", content: "<?php\r\n" + marker + "\r\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(
				filepath.Join(dir, "index.php"), []byte(tc.content), 0o644); err != nil {
				t.Fatalf("write fixture: %v", err)
			}

			config.ProjectDir = ""

			f := NewFingerprint("testCRLF" + tc.name)
			f.Path = dir
			f.Threshold = 1
			f.Entrypoints = []string{"index.php"}
			f.Frameworks = map[string]FrameworkSignature{
				"drupal": {Markers: []string{marker}},
			}

			f.Collect()

			assert.Empty(f.GetErrors())
			assert.Equal(map[string]string{"drupal": "5"}, f.GetData())
		})
	}
}

// TestFileFingerprintUnreadableManifest asserts that a manifest which
// exists but cannot be read is reported as ErrManifestUnreadable, not
// silently scored as zero and not conflated with ErrManifestNotFound -
// the file is present and expected to be readable, so this is a hard
// collection error distinct from "no manifest configured for this
// ecosystem".
//
// The fixture is created at runtime with mode 0000 rather than committed,
// since a permission-stripped file in testdata would not survive checkout
// reliably.
func TestFileFingerprintUnreadableManifest(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root: mode 0000 is still readable, cannot test unreadable file")
	}

	assert := assert.New(t)

	dir := t.TempDir()
	manifest := filepath.Join(dir, "composer.json")
	if err := os.WriteFile(manifest, []byte(`{"require":{"drupal/core-recommended":"^10"}}`), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	if err := os.Chmod(manifest, 0o000); err != nil {
		t.Fatalf("chmod fixture: %v", err)
	}
	t.Cleanup(func() {
		// Restore permissions so t.TempDir cleanup can remove the file.
		_ = os.Chmod(manifest, 0o644)
	})

	config.ProjectDir = ""

	f := NewFingerprint("testUnreadableManifest")
	f.Path = dir
	f.Threshold = 1
	f.Frameworks = map[string]FrameworkSignature{
		"drupal": {Dependencies: []string{"drupal/core-recommended"}},
	}

	f.Collect()

	if assert.NotEmpty(f.GetErrors(), "unreadable manifest must raise a collection error") {
		assert.ErrorIs(f.GetErrors()[0], ErrManifestUnreadable)
	}
	assert.Empty(f.GetData(), "no data should be emitted when collection fails")
}

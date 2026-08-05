package file

import (
	stdjson "encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/theory/jsonpath"

	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

var (
	// ErrNoFrameworks is returned when the plugin is configured without any
	// 'frameworks' signatures to score against.
	ErrNoFrameworks = errors.New(
		"file:fingerprint requires at least one entry under 'frameworks'")
	// ErrNoPath is returned when the plugin is configured without a 'path'
	// to scan.
	ErrNoPath = errors.New("file:fingerprint requires a 'path'")
	// ErrManifestNotFound is returned when at least one configured
	// framework signature uses 'dependencies' but the configured (or
	// default) manifest file does not exist under Path. A missing
	// manifest is reported explicitly rather than silently scoring the
	// dependencies signal as zero, so an operator who expects
	// composer.json to be present (e.g. it was misconfigured, or Path is
	// wrong) is told why the framework was not detected instead of
	// getting a quiet under-score.
	ErrManifestNotFound = errors.New("file:fingerprint manifest not found")
	// ErrManifestUnreadable is returned when the manifest exists but
	// cannot be read (e.g. permissions). Unlike a missing manifest, this
	// is treated as a hard collection error: the file is present and
	// expected to be readable, so silently skipping it risks
	// under-scoring a framework that is actually present.
	ErrManifestUnreadable = errors.New("file:fingerprint manifest unreadable")
	// ErrManifestInvalidJSON is returned when the manifest exists and is
	// readable but cannot be parsed as JSON.
	ErrManifestInvalidJSON = errors.New("file:fingerprint manifest is not valid JSON")
	// ErrInvalidDependencyPath is returned when a configured
	// 'dependency-paths' entry is not a valid RFC 9535 JSONPath
	// expression. jsonpath.Parse is used rather than MustParse so a
	// malformed operator-supplied expression surfaces as a clear error
	// instead of a panic.
	ErrInvalidDependencyPath = errors.New("file:fingerprint invalid dependency-paths expression")
)

// defaultDependencyPaths cover composer's require/require-dev (fixing a
// 0.x gap: utils.HasComposerDependency only ever read 'require') and
// package.json's dependencies/devDependencies, so a single default
// configuration works across the PHP and Node ecosystems without
// operator-supplied 'dependency-paths'.
var defaultDependencyPaths = []string{
	"$.require",
	"$['require-dev']",
	"$.dependencies",
	"$.devDependencies",
}

// FingerprintWeights configures the score contribution of each signal
// type. Markers and Dirs default to 5, matching 0.x's
// sca:application_type (pkg/checks/sca/apptypecheck.go), which added +5
// per marker match and +5 per matching directory. Dependencies defaults
// to 10, matching 0.x's single +10 award for a matched dependency
// (apptypecheck.go:98-102).
type FingerprintWeights struct {
	Markers      int `yaml:"markers"`
	Dirs         int `yaml:"dirs"`
	Dependencies int `yaml:"dependencies"`
}

// FrameworkSignature is the operator-supplied evidence associated with a
// single label. No framework knowledge is baked into the binary - every
// signature (which markers/dirs identify which label) is entirely
// operator config; the binary only supplies the generic matching
// mechanism.
type FrameworkSignature struct {
	// Markers are exact-line strings searched for within each discovered
	// entrypoint file. A marker scores once per (marker, entrypoint file)
	// pair in which it is found, so N markers found across M entrypoint
	// files score weights.markers * N * M - this deliberately reproduces
	// 0.x's per-match accumulation (apptypecheck.go:79-96) rather than
	// capping at one match per marker, so a ported 0.x config scores
	// identically.
	Markers []string `yaml:"markers"`

	// Dirs are directory names searched for anywhere in the tree under
	// Path. Each matching directory occurrence scores independently (so a
	// framework with both "web" and "docroot" present, or the same name
	// nested twice, scores twice), again reproducing 0.x's per-match
	// accumulation.
	Dirs []string `yaml:"dirs"`

	// Dependencies are manifest package names (e.g.
	// "drupal/core-recommended") searched for among the values matched by
	// Fingerprint.DependencyPaths in the manifest. Unlike Markers and
	// Dirs, a matched dependency scores once only, no matter how many
	// configured dependency names match or how many dependency-paths
	// expressions surface them - this reproduces 0.x's single bool-driven
	// +10 award (apptypecheck.go:98-102), the one signal that does not
	// accumulate per-hit.
	Dependencies []string `yaml:"dependencies"`
}

// Fingerprint scores each configured framework signature against the
// filesystem under Path and emits the labels whose accumulated score
// exceeds Threshold. It reproduces the weighted-likelihood model of 0.x's
// sca:application_type check (pkg/checks/sca/apptypecheck.go) rather than
// simple presence matching, so ported 0.x configs (markers/dirs/weights/
// threshold) score identically.
//
// Emits FormatMapString: label -> accumulated score (as a string),
// containing only labels whose score is strictly greater than Threshold.
// Sub-threshold scores are never included in the output - not even as a
// zero or low value - they are logged at debug level only. This is a
// deliberate difference from a raw "all scores" dump: the fact answers
// "is a disallowed framework present?", pairing naturally with the
// detected analyser (any output is a breach), not "what is this app's
// type?", which would need every label's score to compare against an
// expected one.
//
// Threshold default note: 0.x's check code defaults Threshold to 30
// (apptypecheck.go:65-67) but its own reference doc claims 1
// (docs/src/reference/checks/sca-application-type.md:16). This fact
// follows the code (30), which is what 0.x actually executes; the doc
// value was never true at runtime.
//
// Threshold/weights zero-value caveat: because YAML unmarshals an absent
// field to Go's zero value, Threshold: 0 and an unset Weights field are
// indistinguishable from "use the default". An operator who genuinely
// wants "any single match breaches" cannot express Threshold: 0 - they
// must use a threshold of -1 or lower. This is inherited from 0.x, which
// has the identical limitation, and is not fixed here for config
// compatibility with ported 0.x definitions.
//
// Entrypoint matching note: entrypoint files are matched by exact
// basename against Entrypoints, not by 0.x's substring Glob helper
// (utils.Glob does strings.Contains(name, match)), which would also
// match e.g. "myindex.phpx" against "index.php". This is a deliberate
// divergence from 0.x: exact-basename matching is what an operator
// writing `entrypoints: [index.php]` expects, and the substring form
// would inflate marker scores via unintended file matches. A config
// relying on 0.x's looser matching must list the extra basenames
// explicitly.
//
// Data handling note: breach/fact output never includes the matched
// marker text or file path, only the label and its score, so a
// carelessly-written marker cannot echo file content (potentially
// containing secrets) into an audit report. The same applies to the
// dependencies signal: only the label and score are emitted, never the
// matched manifest key or version string.
//
// Dependency-paths safety note: DependencyPaths expressions are parsed
// with jsonpath.Parse (never MustParse), so a malformed operator-supplied
// expression is a collection error, not a panic. RFC 9535 JSONPath has
// no unbounded-recursion construct comparable to e.g. a regex catastrophic
// backtracking - evaluation cost is bounded by the size of the parsed
// manifest document, which this fact already reads fully into memory once
// per Collect call.
type Fingerprint struct {
	fact.BaseFact `yaml:",inline"`

	// Plugin fields.
	Path        string                        `yaml:"path"`
	Threshold   int                           `yaml:"threshold"`
	Entrypoints []string                      `yaml:"entrypoints"`
	Weights     FingerprintWeights            `yaml:"weights"`
	Frameworks  map[string]FrameworkSignature `yaml:"frameworks"`

	// Manifest is the filename (relative to Path) read for the
	// dependencies signal. Defaults to "composer.json". Only a single
	// manifest file is supported per fact instance; a project needing
	// both composer.json and package.json signatures configures two
	// file:fingerprint facts.
	Manifest string `yaml:"manifest"`

	// DependencyPaths are RFC 9535 JSONPath expressions evaluated against
	// the parsed Manifest; every object matched is treated as a map of
	// dependency name to version constraint, and its keys are the
	// candidate dependency names checked against each framework's
	// Dependencies. Defaults to defaultDependencyPaths, covering
	// composer's require/require-dev and package.json's
	// dependencies/devDependencies.
	DependencyPaths []string `yaml:"dependency-paths"`

	// dependencyPaths holds DependencyPaths (or the defaults) parsed once
	// and cached, so a multi-framework Collect does not re-parse the same
	// expressions per framework.
	dependencyPaths []*jsonpath.Path
}

func init() {
	fact.Manager().RegisterFactory("file:fingerprint", func(n string) fact.Facter {
		return NewFingerprint(n)
	})
}

func NewFingerprint(id string) *Fingerprint {
	return &Fingerprint{
		BaseFact: fact.BaseFact{
			BasePlugin: plugin.BasePlugin{
				Id: id,
			},
		},
	}
}

func (p *Fingerprint) GetName() string {
	return "file:fingerprint"
}

func (p *Fingerprint) Collect() {
	contextLogger := log.WithFields(log.Fields{
		"fact-plugin": p.GetName(),
		"fact":        p.GetId(),
	})

	if len(p.Frameworks) == 0 {
		contextLogger.Error("no frameworks configured")
		p.AddErrors(ErrNoFrameworks)
		return
	}

	if p.Path == "" {
		contextLogger.Error("no path configured")
		p.AddErrors(ErrNoPath)
		return
	}

	threshold := p.Threshold
	if threshold == 0 {
		threshold = 30
	}

	weights := p.Weights
	if weights.Markers == 0 {
		weights.Markers = 5
	}
	if weights.Dirs == 0 {
		weights.Dirs = 5
	}
	if weights.Dependencies == 0 {
		weights.Dependencies = 10
	}

	entrypoints := p.Entrypoints
	if len(entrypoints) == 0 {
		entrypoints = []string{"index.php"}
	}

	fullPath := filepath.Join(config.ProjectDir, p.Path)

	contextLogger.WithFields(log.Fields{
		"project-dir": config.ProjectDir,
		"path":        p.Path,
		"threshold":   threshold,
		"weights":     weights,
	}).Debug("fingerprinting codebase")

	// The dependencies signal (manifest read + jsonpath evaluation) is
	// only attempted when at least one configured framework actually
	// uses it - a config that never sets 'dependencies' should not fail
	// just because Path happens not to contain a composer.json.
	needsDependencies := false
	for _, sig := range p.Frameworks {
		if len(sig.Dependencies) > 0 {
			needsDependencies = true
			break
		}
	}

	var dependencyNames map[string]struct{}
	if needsDependencies {
		if err := p.validateDependencyPaths(); err != nil {
			contextLogger.WithError(err).Error("invalid dependency-paths expression")
			p.AddErrors(sentinelError(err))
			return
		}

		names, err := p.readManifestDependencyNames(fullPath)
		if err != nil {
			contextLogger.WithError(err).Error("error reading manifest for dependencies signal")
			p.AddErrors(sentinelError(err))
			return
		}
		dependencyNames = names
	}

	var entrypointFiles []string
	var dirNames []string

	// filepath.WalkDir does not follow symbolic links, so a symlink under
	// Path cannot be used to walk outside it.
	err := filepath.WalkDir(fullPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			dirNames = append(dirNames, d.Name())
			return nil
		}
		if utils.StringSliceContains(entrypoints, d.Name()) {
			entrypointFiles = append(entrypointFiles, path)
		}
		return nil
	})
	if err != nil {
		contextLogger.WithError(err).Error("error walking path")
		p.AddErrors(err)
		return
	}

	// Read each entrypoint once up front rather than re-opening it per
	// (marker, framework) pair. This keeps I/O proportional to the number
	// of entrypoints instead of entrypoints x markers x frameworks, and -
	// more importantly - means an unreadable entrypoint is reported once,
	// as a single collection error, rather than once per marker.
	//
	// An unreadable entrypoint is a hard error, not a skipped file: any
	// marker it might have matched is now invisible, so continuing would
	// under-score the framework and report a framework that is present as
	// absent. Silently under-reporting is the worst failure mode for an
	// audit tool, so this aborts the run (fact errors are fatal in
	// pkg/shipshape/shipshape.go) exactly as an unreadable Path does.
	entrypointLines := make(map[string][]string, len(entrypointFiles))
	for _, ep := range entrypointFiles {
		content, err := os.ReadFile(ep)
		if err != nil {
			contextLogger.WithError(err).WithField("entrypoint", ep).
				Error("error reading entrypoint")
			p.AddErrors(err)
			return
		}
		// Markers are matched against whole lines. Trailing carriage
		// returns are trimmed so a marker matches identically whether the
		// file uses LF or CRLF line endings - bufio.Scanner (used by the
		// 0.x utils.FileContains this replaces) strips them, and a config
		// should not silently stop matching because a file was committed
		// with Windows line endings.
		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			lines[i] = strings.TrimSuffix(line, "\r")
		}
		entrypointLines[ep] = lines
	}

	frameworkNames := make([]string, 0, len(p.Frameworks))
	for name := range p.Frameworks {
		frameworkNames = append(frameworkNames, name)
	}
	sort.Strings(frameworkNames)

	result := map[string]string{}
	for _, name := range frameworkNames {
		sig := p.Frameworks[name]
		score := 0

		for _, marker := range sig.Markers {
			for _, ep := range entrypointFiles {
				if utils.StringSliceContains(entrypointLines[ep], marker) {
					score += weights.Markers
				}
			}
		}

		for _, dirName := range dirNames {
			if utils.StringSliceContains(sig.Dirs, dirName) {
				score += weights.Dirs
			}
		}

		// Dependencies score once only, no matter how many configured
		// names match or how many dependency-paths expressions surface
		// them - reproducing 0.x's single bool-driven award
		// (apptypecheck.go:98-102), the one signal that does not
		// accumulate per-hit.
		for _, dep := range sig.Dependencies {
			if _, ok := dependencyNames[dep]; ok {
				score += weights.Dependencies
				break
			}
		}

		contextLogger.WithFields(log.Fields{
			"framework": name,
			"score":     score,
			"threshold": threshold,
		}).Debug("computed fingerprint score")

		if score > threshold {
			result[name] = strconv.Itoa(score)
		}
	}

	p.Format = data.FormatMapString
	p.SetData(result)
}

// validateDependencyPaths parses DependencyPaths (or
// defaultDependencyPaths, if unset) into p.dependencyPaths, returning an
// error wrapping ErrInvalidDependencyPath if any expression is not a
// valid RFC 9535 JSONPath query. jsonpath.Parse is used rather than
// MustParse so a malformed operator-supplied expression surfaces as a
// clear error instead of a panic. Safe to call more than once: parsing is
// skipped once p.dependencyPaths is populated.
func (p *Fingerprint) validateDependencyPaths() error {
	if p.dependencyPaths != nil {
		return nil
	}

	exprs := p.DependencyPaths
	if len(exprs) == 0 {
		exprs = defaultDependencyPaths
	}

	parsed := make([]*jsonpath.Path, 0, len(exprs))
	for _, expr := range exprs {
		path, err := jsonpath.Parse(expr)
		if err != nil {
			return fmt.Errorf("%w %q: %s", ErrInvalidDependencyPath, expr, err)
		}
		parsed = append(parsed, path)
	}

	p.dependencyPaths = parsed
	return nil
}

// readManifestDependencyNames reads the configured (or default) Manifest
// file under fullPath, evaluates every p.dependencyPaths expression
// against it, and returns the union of keys from every matched object -
// the set of candidate dependency names checked against each framework's
// Dependencies.
//
// A missing manifest, an unreadable manifest, and a manifest that is not
// valid JSON each return a distinct sentinel error (ErrManifestNotFound,
// ErrManifestUnreadable, ErrManifestInvalidJSON respectively) rather than
// a panic or a silent zero score.
func (p *Fingerprint) readManifestDependencyNames(fullPath string) (map[string]struct{}, error) {
	manifestName := p.Manifest
	if manifestName == "" {
		manifestName = "composer.json"
	}

	manifestPath := filepath.Join(fullPath, manifestName)

	if _, err := os.Stat(manifestPath); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("%w: %s", ErrManifestNotFound, manifestPath)
		}
		return nil, fmt.Errorf("%w: %s", ErrManifestUnreadable, err)
	}

	content, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrManifestUnreadable, err)
	}

	var doc any
	if err := stdjson.Unmarshal(content, &doc); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrManifestInvalidJSON, err)
	}

	names := map[string]struct{}{}
	for _, path := range p.dependencyPaths {
		nodes := path.Select(doc)
		for node := range nodes.All() {
			obj, ok := node.(map[string]any)
			if !ok {
				continue
			}
			for name := range obj {
				names[name] = struct{}{}
			}
		}
	}

	return names, nil
}

// sentinelError maps a detailed internal error back to a stable,
// comparable sentinel so downstream reporting is deterministic. Errors
// that do not match a known sentinel are returned unchanged. Mirrors
// json/key.go's sentinel helper.
func sentinelError(err error) error {
	for _, s := range []error{
		ErrManifestNotFound,
		ErrManifestUnreadable,
		ErrManifestInvalidJSON,
		ErrInvalidDependencyPath,
	} {
		if errors.Is(err, s) {
			return s
		}
	}
	return err
}

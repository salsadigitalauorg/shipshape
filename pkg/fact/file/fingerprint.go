package file

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

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
)

// FingerprintWeights configures the score contribution of each signal
// type. Markers and Dirs default to 5, matching 0.x's
// sca:application_type (pkg/checks/sca/apptypecheck.go), which added +5
// per marker match and +5 per matching directory. The dependencies signal
// (weight, default 10) is added in a later slice.
type FingerprintWeights struct {
	Markers int `yaml:"markers"`
	Dirs    int `yaml:"dirs"`
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
// containing secrets) into an audit report.
type Fingerprint struct {
	fact.BaseFact `yaml:",inline"`

	// Plugin fields.
	Path        string                        `yaml:"path"`
	Threshold   int                           `yaml:"threshold"`
	Entrypoints []string                      `yaml:"entrypoints"`
	Weights     FingerprintWeights            `yaml:"weights"`
	Frameworks  map[string]FrameworkSignature `yaml:"frameworks"`
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

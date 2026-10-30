// Command brewgen renders Homebrew formulae for the shipshape tap from a
// GoReleaser checksums.txt file. It exists mainly to produce multiple
// versioned formulae.
package main

import (
	"bufio"
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
)

//go:embed templates/*.tmpl
var templates embed.FS

// Platform identifies one published archive by its GoReleaser name_template
// output, e.g. "shipshape-Darwin-aarch64.tar.gz".
type Platform struct {
	// OS as GoReleaser titles it in the archive name.
	OS string
	// Arch as GoReleaser renames it in the archive name.
	Arch string
}

// platforms are the four archives the tap formulae depend on. The order is
// fixed so rendered output is deterministic.
var platforms = []Platform{
	{OS: "Darwin", Arch: "x86_64"},
	{OS: "Darwin", Arch: "aarch64"},
	{OS: "Linux", Arch: "x86_64"},
	{OS: "Linux", Arch: "aarch64"},
}

// Binding is a resolved platform: the archive name, its download URL and its
// SHA256, ready for the template.
type Binding struct {
	OS     string
	Arch   string
	Name   string
	URL    string
	SHA256 string
}

// IsMacOS reports whether this binding belongs in the on_macos block.
func (b Binding) IsMacOS() bool { return b.OS == "Darwin" }

// IsIntel reports whether this binding belongs in the on_intel block.
func (b Binding) IsIntel() bool { return b.Arch == "x86_64" }

// Formula is the template payload for a single .rb file.
type Formula struct {
	// ClassName is the Ruby class name Homebrew derives from the filename.
	ClassName string
	// Name is the installed binary name, e.g. "shipshape".
	Name string
	// Version is the release version with any leading "v" stripped.
	Version string
	// Tag is the git tag, retained for building download URLs.
	Tag string
	// Repo is the "owner/name" GitHub repository.
	Repo string
	// Branch is the git branch a --HEAD build compiles from.
	Branch string
	// Desc and Homepage populate the matching formula stanzas.
	Desc     string
	Homepage string
	License  string
	// Bindings are the four resolved platforms.
	Bindings []Binding
	// Versioned marks a pinned formula, which omits the head block and is
	// keg_only so it does not collide with the main formula's symlinks.
	Versioned bool
}

// MacOS returns the bindings belonging in the on_macos block.
func (f Formula) MacOS() []Binding { return f.filter(true) }

// Linux returns the bindings belonging in the on_linux block.
func (f Formula) Linux() []Binding { return f.filter(false) }

func (f Formula) filter(macos bool) []Binding {
	var out []Binding
	for _, b := range f.Bindings {
		if b.IsMacOS() == macos {
			out = append(out, b)
		}
	}
	return out
}

// checksumLine matches a GoReleaser checksums.txt entry: a hex digest, run of
// whitespace, then the artifact filename.
var checksumLine = regexp.MustCompile(`^([0-9a-fA-F]{64})\s+(\S+)$`)

// ParseChecksums reads a GoReleaser checksums.txt and returns filename->SHA256.
func ParseChecksums(r *bufio.Scanner) (map[string]string, error) {
	sums := map[string]string{}
	line := 0
	for r.Scan() {
		line++
		text := strings.TrimSpace(r.Text())
		if text == "" {
			continue
		}
		m := checksumLine.FindStringSubmatch(text)
		if m == nil {
			return nil, fmt.Errorf("checksums line %d is malformed: %q", line, text)
		}
		sums[m[2]] = strings.ToLower(m[1])
	}
	if err := r.Err(); err != nil {
		return nil, fmt.Errorf("reading checksums: %w", err)
	}
	if len(sums) == 0 {
		return nil, fmt.Errorf("checksums file contained no entries")
	}
	return sums, nil
}

// ClassS reproduces Homebrew's Formulary.class_s, which derives a formula's
// Ruby class name from its file name.
//
// Ported from Library/Homebrew/formulary.rb:
//
//	class_name = name.capitalize
//	class_name.gsub!(/[-_.\s][a-zA-Z0-9]/) { |m| m.chars.fetch(-1).upcase }
//	class_name.tr!("+", "x")
//	class_name.sub!(/(.)@(\d)/, "\\1AT\\2")
//
// A mismatch between the class name and the file name makes the formula
// unloadable, so this must track Homebrew exactly.
var (
	classSSeparator = regexp.MustCompile(`[-_.\s][a-zA-Z0-9]`)
	classSAt        = regexp.MustCompile(`(.)@(\d)`)
)

func ClassS(name string) string {
	// Ruby's String#capitalize upcases the first character and downcases the
	// rest; Go has no direct equivalent.
	className := name
	if className != "" {
		className = strings.ToUpper(className[:1]) + strings.ToLower(className[1:])
	}

	className = classSSeparator.ReplaceAllStringFunc(className, func(m string) string {
		// Drop the separator, upcase the character that followed it.
		return strings.ToUpper(m[len(m)-1:])
	})

	className = strings.ReplaceAll(className, "+", "x")

	// Only the first @-before-digit is rewritten, matching Ruby's sub!.
	if loc := classSAt.FindStringSubmatchIndex(className); loc != nil {
		className = className[:loc[2]] + className[loc[2]:loc[3]] + "AT" +
			className[loc[4]:loc[5]] + className[loc[5]:]
	}

	return className
}

// Options configure a Render call.
type Options struct {
	// Tag is the release tag, e.g. "v1.0.0-alpha.1.6".
	Tag string
	// Repo is "owner/name" on GitHub.
	Repo string
	// Name is the formula name, e.g. "shipshape".
	Name string
	// Branch is the branch a --HEAD build compiles from.
	Branch string
	Desc   string
	// Homepage defaults to the GitHub repository URL when empty.
	Homepage string
	License  string
}

// Rendered is one formula file awaiting write.
type Rendered struct {
	Filename string
	Content  []byte
}

// Render produces the main formula and the pinned versioned formula for a
// release. It fails if any of the four expected archives is absent from sums:
// a formula with an empty sha256 fails opaquely at install time, so it is
// better to break the release than to publish one.
func Render(opts Options, sums map[string]string) ([]Rendered, error) {
	if opts.Tag == "" {
		return nil, fmt.Errorf("tag is required")
	}
	if opts.Repo == "" {
		return nil, fmt.Errorf("repo is required")
	}
	if opts.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	version := strings.TrimPrefix(opts.Tag, "v")
	homepage := opts.Homepage
	if homepage == "" {
		homepage = "https://github.com/" + opts.Repo
	}

	bindings := make([]Binding, 0, len(platforms))
	var missing []string
	for _, p := range platforms {
		name := fmt.Sprintf("%s-%s-%s.tar.gz", opts.Name, p.OS, p.Arch)
		sha, ok := sums[name]
		if !ok {
			missing = append(missing, name)
			continue
		}
		bindings = append(bindings, Binding{
			OS:   p.OS,
			Arch: p.Arch,
			Name: name,
			URL: fmt.Sprintf("https://github.com/%s/releases/download/%s/%s",
				opts.Repo, opts.Tag, name),
			SHA256: sha,
		})
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return nil, fmt.Errorf("checksums file is missing required archives: %s",
			strings.Join(missing, ", "))
	}

	base := Formula{
		Name:     opts.Name,
		Version:  version,
		Tag:      opts.Tag,
		Repo:     opts.Repo,
		Branch:   opts.Branch,
		Desc:     opts.Desc,
		Homepage: homepage,
		License:  opts.License,
		Bindings: bindings,
	}

	tmpl, err := template.ParseFS(templates, "templates/*.tmpl")
	if err != nil {
		return nil, fmt.Errorf("parsing templates: %w", err)
	}

	out := make([]Rendered, 0, 2)

	// Main formula: tracks latest, offers a --HEAD source build.
	mainFile := opts.Name + ".rb"
	main := base
	main.ClassName = ClassS(strings.TrimSuffix(mainFile, ".rb"))
	content, err := execute(tmpl, "formula.rb.tmpl", main)
	if err != nil {
		return nil, err
	}
	out = append(out, Rendered{Filename: mainFile, Content: content})

	// Pinned formula: exact version, binary only, keg_only.
	pinnedFile := fmt.Sprintf("%s@%s.rb", opts.Name, version)
	pinned := base
	pinned.Versioned = true
	pinned.ClassName = ClassS(strings.TrimSuffix(pinnedFile, ".rb"))
	content, err = execute(tmpl, "formula.rb.tmpl", pinned)
	if err != nil {
		return nil, err
	}
	out = append(out, Rendered{Filename: pinnedFile, Content: content})

	return out, nil
}

func execute(tmpl *template.Template, name string, data Formula) ([]byte, error) {
	buf := &bytes.Buffer{}
	if err := tmpl.ExecuteTemplate(buf, name, data); err != nil {
		return nil, fmt.Errorf("rendering %s: %w", name, err)
	}
	return buf.Bytes(), nil
}

// Write writes rendered formulae into dir, creating it if needed. Unlike
// cmd/gen's createFile, existing files are overwritten: a formula must be
// replaced on every release.
func Write(dir string, files []Rendered) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", dir, err)
	}
	for _, f := range files {
		path := filepath.Join(dir, f.Filename)
		if err := os.WriteFile(path, f.Content, 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}
	}
	return nil
}

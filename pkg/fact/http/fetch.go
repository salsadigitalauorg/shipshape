package http

import (
	"errors"
	"fmt"
	"net/url"

	log "github.com/sirupsen/logrus"

	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

var (
	// ErrNoUrl is returned when the plugin is configured without a url.
	ErrNoUrl = errors.New("http:fetch requires a 'url'")
	// ErrInvalidUrl is returned when the url cannot be parsed.
	ErrInvalidUrl = errors.New("http:fetch requires a valid, absolute url")
	// ErrSchemeNotAllowed is returned when the url scheme is not https, and
	// allow-insecure has not been set.
	ErrSchemeNotAllowed = errors.New(
		"http:fetch requires an https url; set allow-insecure to permit http")
)

// Fetch retrieves the content of a remote URL over HTTPS and emits it as raw
// bytes. It is a leaf collector (no input), analogous to file:read but for
// remote sources - e.g. fetching a template file from a GitHub repository to
// compare against a provisioned project's copy.
//
// HTTPS is required by default; ISM-1139 recommends encrypting data in
// transit, and defaulting to plain HTTP for an operator-supplied URL risks
// silently sending audit traffic unencrypted. Set AllowInsecure to opt out
// for local/test scenarios.
//
// Only unauthenticated, publicly accessible URLs are supported. Fetching a
// private repository requires a locally-cloned copy read via file:read; a
// 404 response is a common symptom and includes a hint to that effect,
// because GitHub returns 404 (not 401/403) for unauthorised private repos.
type Fetch struct {
	fact.BaseFact `yaml:",inline"`

	// Plugin fields.
	Url           string `yaml:"url"`
	AllowInsecure bool   `yaml:"allow-insecure"`
}

//go:generate go run ../../../cmd/gen.go fact-plugin --package=http

func init() {
	fact.Manager().RegisterFactory("http:fetch", func(n string) fact.Facter {
		return NewFetch(n)
	})
}

func NewFetch(id string) *Fetch {
	return &Fetch{
		BaseFact: fact.BaseFact{
			BasePlugin: plugin.BasePlugin{
				Id: id,
			},
		},
	}
}

func (p *Fetch) GetName() string {
	return "http:fetch"
}

func (p *Fetch) Collect() {
	contextLogger := log.WithFields(log.Fields{
		"fact-plugin": p.GetName(),
		"fact":        p.GetId(),
	})

	if p.Url == "" {
		p.AddErrors(ErrNoUrl)
		return
	}

	parsed, err := url.Parse(p.Url)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		p.AddErrors(ErrInvalidUrl)
		return
	}
	if parsed.Scheme != "https" && !p.AllowInsecure {
		p.AddErrors(ErrSchemeNotAllowed)
		return
	}

	contextLogger.WithField("url", p.Url).Debug("fetching remote content")

	content, err := utils.FetchContentFromUrl(p.Url)
	if err != nil {
		contextLogger.WithError(err).Error("error fetching url")
		p.AddErrors(fmt.Errorf(
			"fetching %s failed: %w (if this is a private repository, "+
				"clone it locally and use file:read instead - GitHub returns "+
				"404 for unauthorised private repos, not 401/403)", p.Url, err))
		return
	}

	p.Format = data.FormatRaw
	p.SetData(content)
}

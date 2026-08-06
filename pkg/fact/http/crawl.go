package http

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"golang.org/x/net/html"

	log "github.com/sirupsen/logrus"

	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/env"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	"github.com/salsadigitalauorg/shipshape/pkg/plugin"
)

var (
	// ErrNoCrawlUrl is returned when the plugin is configured without a url.
	ErrNoCrawlUrl = errors.New("http:crawl requires a 'url'")
	// ErrInvalidCrawlUrl is returned when the url cannot be parsed.
	ErrInvalidCrawlUrl = errors.New(
		"http:crawl requires a valid, absolute url")
	// ErrCrawlSchemeNotAllowed is returned when the url scheme is not
	// https, and allow-insecure has not been set.
	ErrCrawlSchemeNotAllowed = errors.New(
		"http:crawl requires an https url; set allow-insecure to permit http")
	// ErrInvalidCrawlTimeout is returned when timeout cannot be parsed as a
	// duration.
	ErrInvalidCrawlTimeout = errors.New(
		"http:crawl timeout must be a valid duration, e.g. '30s'")
	// ErrInvalidCrawlLimit is returned when limit is negative.
	ErrInvalidCrawlLimit = errors.New(
		"http:crawl limit must be zero (for the default) or a positive number")
	// ErrInvalidCrawlMaxDepth is returned when max-depth is negative.
	ErrInvalidCrawlMaxDepth = errors.New(
		"http:crawl max-depth must be zero (for the default) or a positive number")
	// ErrOffAllowlistRedirect is returned when a response redirects to a
	// host outside the root host + include-domains allowlist. The
	// redirect is refused before it is issued - ISM-1288 / CWE-918: a
	// compromised or malicious audited site must not be able to steer
	// audit traffic at arbitrary third-party infrastructure. A refusal on
	// a discovered link is logged and skipped, not recorded as a breach -
	// many sites legitimately redirect off-host (SSO, www/apex
	// normalisation, link-tracking wrappers). A refusal on the crawl root
	// itself is a hard error - see ErrRootRedirectOffAllowlist.
	ErrOffAllowlistRedirect = errors.New(
		"http:crawl refused a redirect to a non-allow-listed host")
	// ErrRootRedirectOffAllowlist is returned when the configured url
	// itself redirects to a host outside the allowlist. Without this
	// check the crawl would silently visit nothing and report success.
	ErrRootRedirectOffAllowlist = errors.New(
		"http:crawl url redirects to a host outside the allowlist; " +
			"set url to the post-redirect host, or add the target to include-domains")
)

// defaultCrawlMaxDepth bounds link-following depth from the root url when
// max-depth is not configured. The root url itself is depth 0.
const defaultCrawlMaxDepth = 3

// defaultCrawlLimit bounds the total number of requests issued when limit
// is not configured. ISM-1288: an unbounded crawl of operator-supplied
// infrastructure risks an unintentional denial of service against the
// audited site; a crawl that needs a wider budget must opt in explicitly.
const defaultCrawlLimit = 100

// defaultCrawlTimeout is the per-request timeout used when timeout is not
// configured, matching utils.fetchTimeout (pkg/utils/utils.go).
const defaultCrawlTimeout = 30 * time.Second

// maxCrawlResponseSize bounds the number of bytes read from each crawled
// response body, matching utils.maxFetchResponseSize
// (pkg/utils/utils.go), applied per-request rather than once since a crawl
// issues many requests.
const maxCrawlResponseSize = 10 * 1024 * 1024 // 10 MiB

// maxCrawlRedirects caps the number of redirects followed per request,
// matching net/http's unexported defaultCheckRedirect. Supplying a custom
// CheckRedirect (required to enforce the domain allowlist on every hop)
// replaces that default entirely, so the cap must be reimplemented here -
// omitting it would trade the allowlist bypass this guards against for an
// unbounded redirect loop.
const maxCrawlRedirects = 10

// Crawl walks a site starting from a root url, following same-site (or
// explicitly allow-listed) links up to a bounded depth and request count,
// and emits the HTTP status (or transport error) of every non-2xx or
// failed request encountered. It is a leaf collector (no input), analogous
// to http:fetch but for auditing a site's link graph rather than fetching
// a single resource - e.g. detecting broken links across a site.
//
// HTTPS is required by default; ISM-1139 recommends encrypting data in
// transit, and defaulting to plain HTTP for an operator-supplied url risks
// silently auditing unencrypted traffic. Set AllowInsecure to opt out for
// local/test scenarios.
//
// Only the root host, plus any hosts in IncludeDomains, are followed -
// matched as an exact host string, so subdomains and www/apex variants of
// the same site are not implied. A link to any other host is skipped
// entirely: not visited, not recorded, not even as a breach. A redirect to
// any other host is refused before it is issued - enforced on every hop,
// not only the initial request - and is logged rather than recorded, since
// sites commonly redirect off-host for legitimate reasons (SSO, www/apex
// normalisation, link-tracking wrappers); add the target to IncludeDomains
// to have it crawled instead. The one exception is Url itself: if the root
// redirects off-allowlist the crawl fails outright
// (ErrRootRedirectOffAllowlist), since silently visiting nothing and
// reporting success would be worse than an explicit error.
//
// Traversal is sequential (single goroutine), which keeps runs
// deterministic and avoids inadvertently load-testing the audited site.
type Crawl struct {
	fact.BaseFact       `yaml:",inline"`
	env.BaseEnvResolver `yaml:",inline"`

	// Plugin fields.
	Url            string   `yaml:"url"`
	IncludeDomains []string `yaml:"include-domains"`
	IncludeUrls    []string `yaml:"include-urls"`
	MaxDepth       int      `yaml:"max-depth"`
	Limit          int      `yaml:"limit"`
	Timeout        string   `yaml:"timeout"`
	AllowInsecure  bool     `yaml:"allow-insecure"`
}

//go:generate go run ../../../cmd/gen.go fact-plugin --package=http

func init() {
	fact.Manager().RegisterFactory("http:crawl", func(n string) fact.Facter {
		return NewCrawl(n)
	})
}

func NewCrawl(id string) *Crawl {
	return &Crawl{
		BaseFact: fact.BaseFact{
			BasePlugin: plugin.BasePlugin{
				Id: id,
			},
		},
	}
}

func (p *Crawl) GetName() string {
	return "http:crawl"
}

// queueItem is one url pending a crawl request. isRoot marks the crawl's
// original url (as opposed to an include-urls seed or a discovered link,
// both of which also start at depth 0) - see the isRoot check in Collect
// for why the distinction matters.
type queueItem struct {
	u      *url.URL
	depth  int
	isRoot bool
}

// normalizeCrawlURL returns a copy of u with its fragment cleared -
// fragments are resolved client-side and never affect what the server
// returns, so keeping them would visit the same resource under multiple
// visited-set keys - and an empty path normalised to "/", so that
// "http://host" and "http://host/" are treated as the same resource and
// the root is never queued (and counted against limit) twice.
func normalizeCrawlURL(u *url.URL) *url.URL {
	c := *u
	c.Fragment = ""
	if c.Path == "" {
		c.Path = "/"
	}
	return &c
}

func (p *Crawl) Collect() {
	contextLogger := log.WithFields(log.Fields{
		"fact-plugin": p.GetName(),
		"fact":        p.GetId(),
	})

	envMap, err := p.GetEnvMap()
	if err != nil {
		p.AddErrors(err)
		return
	}

	rawUrl, err := env.ResolveValue(envMap, p.Url)
	if err != nil {
		p.AddErrors(err)
		return
	}

	if rawUrl == "" {
		p.AddErrors(ErrNoCrawlUrl)
		return
	}

	root, err := url.Parse(rawUrl)
	if err != nil || root.Scheme == "" || root.Host == "" {
		p.AddErrors(ErrInvalidCrawlUrl)
		return
	}
	if root.Scheme != "https" && !p.AllowInsecure {
		p.AddErrors(ErrCrawlSchemeNotAllowed)
		return
	}

	timeout := defaultCrawlTimeout
	if p.Timeout != "" {
		timeout, err = time.ParseDuration(p.Timeout)
		if err != nil {
			p.AddErrors(ErrInvalidCrawlTimeout)
			return
		}
	}

	if p.MaxDepth < 0 {
		p.AddErrors(ErrInvalidCrawlMaxDepth)
		return
	}
	maxDepth := p.MaxDepth
	if maxDepth == 0 {
		maxDepth = defaultCrawlMaxDepth
	}

	if p.Limit < 0 {
		p.AddErrors(ErrInvalidCrawlLimit)
		return
	}
	limit := p.Limit
	if limit == 0 {
		limit = defaultCrawlLimit
	}

	allowedHosts := map[string]bool{root.Host: true}
	for _, d := range p.IncludeDomains {
		allowedHosts[d] = true
	}

	root = normalizeCrawlURL(root)

	// A custom CheckRedirect is required to enforce the allowlist on every
	// redirect hop, not just the initial request - net/http otherwise
	// follows redirects (subject only to its own default 10-redirect cap,
	// which is disabled the moment CheckRedirect is set, hence
	// maxCrawlRedirects reimplementing it below) with no host check at all.
	client := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if !allowedHosts[req.URL.Host] {
				return fmt.Errorf("%w: %s", ErrOffAllowlistRedirect, req.URL.Host)
			}
			if len(via) >= maxCrawlRedirects {
				return fmt.Errorf("stopped after %d redirects", maxCrawlRedirects)
			}
			return nil
		},
	}

	visited := map[string]bool{}
	var queue []queueItem
	queue = append(queue, queueItem{u: root, depth: 0, isRoot: true})
	visited[root.String()] = true

	for _, raw := range p.IncludeUrls {
		u, err := url.Parse(raw)
		if err != nil {
			contextLogger.WithField("url", raw).WithError(err).
				Warn("skipping invalid include-urls entry")
			continue
		}
		resolved := normalizeCrawlURL(root.ResolveReference(u))
		if visited[resolved.String()] {
			continue
		}
		visited[resolved.String()] = true
		queue = append(queue, queueItem{u: resolved, depth: 0})
	}

	results := map[string]string{}
	requests := 0

	for len(queue) > 0 && requests < limit {
		item := queue[0]
		queue = queue[1:]
		requests++

		links, status, err := fetchAndExtractLinks(client, item.u)
		if err != nil {
			if errors.Is(err, ErrOffAllowlistRedirect) {
				if item.isRoot {
					p.AddErrors(fmt.Errorf("%w (%s)", ErrRootRedirectOffAllowlist, err))
					return
				}
				contextLogger.WithField("from", item.u.String()).WithError(err).
					Warn("refused redirect to non-allow-listed host; not followed, not recorded")
				continue
			}
			contextLogger.WithField("url", item.u.String()).WithError(err).
				Debug("crawl request failed")
			results[item.u.String()] = "error: " + err.Error()
			continue
		}
		if status < 200 || status >= 300 {
			results[item.u.String()] = strconv.Itoa(status)
		}
		if status < 200 || status >= 300 || item.depth >= maxDepth {
			continue
		}

		for _, link := range links {
			resolved := normalizeCrawlURL(item.u.ResolveReference(link))
			if resolved.Scheme != "http" && resolved.Scheme != "https" {
				continue
			}
			if !allowedHosts[resolved.Host] {
				continue
			}
			key := resolved.String()
			if visited[key] {
				continue
			}
			visited[key] = true
			queue = append(queue, queueItem{u: resolved, depth: item.depth + 1})
		}
	}

	p.Format = data.FormatMapString
	p.SetData(results)
}

// fetchAndExtractLinks issues a single bounded GET request to u using the
// given client and, if the response is 2xx HTML (or XHTML), extracts and
// returns the raw href values of every anchor tag found. It returns the
// response status code (or 0 if the request failed outright) and any
// transport-level error. client is shared across the whole crawl so
// connections are reused, and its CheckRedirect enforces the domain
// allowlist and a redirect-count cap on every hop - see Collect.
func fetchAndExtractLinks(client *http.Client, u *url.URL) ([]*url.URL, int, error) {
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, 0, err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer rsp.Body.Close()

	status := rsp.StatusCode
	if status < 200 || status >= 300 {
		return nil, status, nil
	}

	// mime.ParseMediaType lowercases the type and strips parameters (e.g.
	// charset), so "TEXT/HTML" and "text/html; charset=utf-8" both match.
	// Its error is deliberately discarded: for some malformed-but-still
	// meaningful inputs (e.g. a trailing ";;") it returns a non-nil error
	// alongside a perfectly usable mediatype, and an empty header also
	// errors - gating on the error would silently drop pages a stricter
	// check would still correctly classify.
	mediaType, _, _ := mime.ParseMediaType(rsp.Header.Get("Content-Type"))
	switch mediaType {
	case "text/html", "application/xhtml+xml":
	default:
		return nil, status, nil
	}

	limited := io.LimitReader(rsp.Body, maxCrawlResponseSize+1)
	links := extractLinks(limited)

	return links, status, nil
}

// extractLinks scans r for anchor tags and returns the raw (unresolved)
// href values found, in document order. Malformed HTML is tolerated -
// html.Tokenizer stops at the first error/EOF and whatever was parsed up
// to that point is returned, matching the resilience expected of an audit
// tool crawling third-party markup it does not control.
func extractLinks(r io.Reader) []*url.URL {
	var links []*url.URL
	z := html.NewTokenizer(r)

	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			return links
		}
		if tt != html.StartTagToken && tt != html.SelfClosingTagToken {
			continue
		}

		tok := z.Token()
		if tok.Data != "a" {
			continue
		}

		for _, attr := range tok.Attr {
			if attr.Key != "href" || attr.Val == "" {
				continue
			}
			href, err := url.Parse(attr.Val)
			if err != nil {
				continue
			}
			links = append(links, href)
		}
	}
}

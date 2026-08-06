package http_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	. "github.com/salsadigitalauorg/shipshape/pkg/fact/http"
)

func TestCrawlInit(t *testing.T) {
	assert := assert.New(t)

	factPlugin := fact.Manager().GetFactories()["http:crawl"]("testCrawl")
	assert.NotNil(factPlugin)
	crawlFacter, ok := factPlugin.(*Crawl)
	assert.True(ok)
	assert.Equal("testCrawl", crawlFacter.GetId())
}

func TestCrawlPluginName(t *testing.T) {
	assert.Equal(t, "http:crawl", NewCrawl("testCrawl").GetName())
}

func TestCrawlNoUrl(t *testing.T) {
	c := NewCrawl("testCrawl")
	c.AllowInsecure = true
	c.Collect()
	assert.Equal(t, []error{ErrNoCrawlUrl}, c.GetErrors())
}

func TestCrawlInvalidUrl(t *testing.T) {
	c := NewCrawl("testCrawl")
	c.Url = "not-a-url"
	c.AllowInsecure = true
	c.Collect()
	assert.Equal(t, []error{ErrInvalidCrawlUrl}, c.GetErrors())
}

func TestCrawlInsecureSchemeRejectedByDefault(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("should not be fetched"))
	}))
	defer svr.Close()

	c := NewCrawl("testCrawl")
	c.Url = svr.URL // httptest.NewServer is plain http://
	c.Collect()
	assert.Equal(t, []error{ErrCrawlSchemeNotAllowed}, c.GetErrors())
}

func TestCrawlInvalidTimeout(t *testing.T) {
	c := NewCrawl("testCrawl")
	c.Url = "https://example.org"
	c.Timeout = "not-a-duration"
	c.Collect()
	assert.Equal(t, []error{ErrInvalidCrawlTimeout}, c.GetErrors())
}

// TestCrawlStatusBoundaries is the regression guard for colly's
// handleOnError treating StatusCode < 203 as success, which misreports 203
// and 204 as errors (see docs/plans/2026-08-06-slice4-http-crawl.md §1).
func TestCrawlStatusBoundaries(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<a href="/no-content">no content</a>`))
	})
	mux.HandleFunc("/no-content", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent) // 204
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()

	c := NewCrawl("testCrawl")
	c.Url = svr.URL
	c.AllowInsecure = true
	c.Collect()

	assert.Empty(t, c.GetErrors())
	assert.Equal(t, data.FormatMapString, c.GetFormat())
	assert.Empty(t, data.AsMapString(c.GetData()),
		"204 must not be recorded as a breach")
}

// TestCrawl203NotRecorded asserts 203 is treated as success, not recorded
// in the output - the same regression guard as TestCrawlStatusBoundaries's
// 204 case, since colly's bug misclassifies both 203 and 204 as errors
// (docs/plans/2026-08-06-slice4-http-crawl.md §1). 203 is a 2xx code; this
// design reports only non-2xx statuses and transport errors.
func TestCrawl203NotRecorded(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<a href="/partial">partial</a>`))
	})
	mux.HandleFunc("/partial", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNonAuthoritativeInfo) // 203
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()

	c := NewCrawl("testCrawl")
	c.Url = svr.URL
	c.AllowInsecure = true
	c.Collect()

	assert.Empty(t, c.GetErrors())
	assert.Empty(t, data.AsMapString(c.GetData()),
		"203 must not be recorded as a breach")
}

func TestCrawl404Recorded(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<a href="/missing">missing</a>`))
	})
	mux.HandleFunc("/missing", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()

	c := NewCrawl("testCrawl")
	c.Url = svr.URL
	c.AllowInsecure = true
	c.Collect()

	assert.Empty(t, c.GetErrors())
	results := data.AsMapString(c.GetData())
	assert.Equal(t, "404", results[svr.URL+"/missing"])
	assert.Len(t, results, 1)
}

func TestCrawlHealthySiteEmitsEmptyMap(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<a href="/about">about</a>`))
	})
	mux.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`no links here`))
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()

	c := NewCrawl("testCrawl")
	c.Url = svr.URL
	c.AllowInsecure = true
	c.Collect()

	assert.Empty(t, c.GetErrors())
	assert.Empty(t, data.AsMapString(c.GetData()))
}

// TestCrawlMaxDepthOffByOne asserts the root url is depth 0: with
// max-depth: 1, links found on the root page are fetched (depth 1), but
// links found on THOSE pages are not (would be depth 2).
func TestCrawlMaxDepthOffByOne(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<a href="/depth1">depth1</a>`))
	})
	mux.HandleFunc("/depth1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<a href="/missing">missing</a>`))
	})
	mux.HandleFunc("/missing", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()

	c := NewCrawl("testCrawl")
	c.Url = svr.URL
	c.AllowInsecure = true
	c.MaxDepth = 1
	c.Collect()

	assert.Empty(t, c.GetErrors())
	results := data.AsMapString(c.GetData())
	assert.Empty(t, results,
		"/missing is at depth 2 and must not be reached when max-depth is 1")
}

func TestCrawlLimitBoundsTotalRequests(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<a href="/a">a</a><a href="/b">b</a><a href="/c">c</a>`))
	})
	for _, path := range []string{"/a", "/b", "/c"} {
		p := path
		mux.HandleFunc(p, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})
	}
	svr := httptest.NewServer(mux)
	defer svr.Close()

	c := NewCrawl("testCrawl")
	c.Url = svr.URL
	c.AllowInsecure = true
	c.Limit = 2
	c.Collect()

	assert.Empty(t, c.GetErrors())
	results := data.AsMapString(c.GetData())
	assert.Len(t, results, 1,
		"only 1 of the 3 linked pages fits within the root + limit=2 budget")
}

// TestCrawlAllowlistBypassSchemeRelative guards against a scheme-relative
// href (//evil.example/x) being resolved to a host outside the allowlist
// yet still followed because its Host looks unrelated to the check.
func TestCrawlAllowlistBypassSchemeRelative(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<a href="//evil.example/x">evil</a>`))
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()

	c := NewCrawl("testCrawl")
	c.Url = svr.URL
	c.AllowInsecure = true
	c.Collect()

	assert.Empty(t, c.GetErrors())
	assert.Empty(t, data.AsMapString(c.GetData()),
		"scheme-relative link to a non-allow-listed host must not be followed")
}

// TestCrawlAllowlistBypassUserinfo guards against a link whose userinfo
// component embeds the allowed host (http://ok.com@evil.com/) being
// mistaken for the allowed host; url.URL.Host never includes userinfo, so
// this is mostly a documentation-by-test of that guarantee.
func TestCrawlAllowlistBypassUserinfo(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		host := r.Host
		w.Write([]byte(`<a href="http://` + host + `@evil.example/x">evil</a>`))
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()

	c := NewCrawl("testCrawl")
	c.Url = svr.URL
	c.AllowInsecure = true
	c.Collect()

	assert.Empty(t, c.GetErrors())
	assert.Empty(t, data.AsMapString(c.GetData()),
		"userinfo-embedded host must not bypass the allowlist")
}

func TestCrawlIncludeDomainsExtendsAllowlist(t *testing.T) {
	mux2 := http.NewServeMux()
	mux2.HandleFunc("/other", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	svr2 := httptest.NewServer(mux2)
	defer svr2.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<a href="` + svr2.URL + `/other">other</a>`))
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()

	c := NewCrawl("testCrawl")
	c.Url = svr.URL
	c.AllowInsecure = true
	c.IncludeDomains = []string{strings.TrimPrefix(svr2.URL, "http://")}
	c.Collect()

	assert.Empty(t, c.GetErrors())
	results := data.AsMapString(c.GetData())
	assert.Equal(t, "404", results[svr2.URL+"/other"])
}

func TestCrawlIncludeUrlsSeededRegardlessOfDiscovery(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`no links here`))
	})
	mux.HandleFunc("/orphan", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()

	c := NewCrawl("testCrawl")
	c.Url = svr.URL
	c.AllowInsecure = true
	c.IncludeUrls = []string{"/orphan"}
	c.Collect()

	assert.Empty(t, c.GetErrors())
	results := data.AsMapString(c.GetData())
	assert.Equal(t, "404", results[svr.URL+"/orphan"])
}

func TestCrawlTransportErrorRecorded(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<a href="http://127.0.0.1:1/unreachable">unreachable</a>`))
	}))
	defer svr.Close()

	c := NewCrawl("testCrawl")
	c.Url = svr.URL
	c.AllowInsecure = true
	c.IncludeDomains = []string{"127.0.0.1:1"}
	c.Collect()

	assert.Empty(t, c.GetErrors())
	results := data.AsMapString(c.GetData())
	found := false
	for u, v := range results {
		if strings.Contains(u, "127.0.0.1:1") {
			found = true
			assert.True(t, strings.HasPrefix(v, "error: "))
		}
	}
	assert.True(t, found, "unreachable host must be recorded as an error")
}

func TestCrawlResolveEnvUsesDotEnvFile(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()

	dir := t.TempDir()
	envFile := dir + "/.env"
	assert.NoError(t, os.WriteFile(envFile, []byte("CRAWL_BASE_URL="+svr.URL+"\n"), 0o600))

	c := NewCrawl("testCrawl")
	c.Url = "${CRAWL_BASE_URL}"
	c.AllowInsecure = true
	c.ResolveEnv = true
	c.EnvFile = envFile
	c.Collect()

	assert.Empty(t, c.GetErrors())
	results := data.AsMapString(c.GetData())
	assert.Equal(t, "404", results[svr.URL+"/"])
}

// TestCrawlRefusesOffAllowlistRedirect is the key regression guard for the
// allowlist-bypass-via-redirect vulnerability: an on-allowlist page
// redirects to an off-allowlist host, and the offsite server must never be
// contacted at all - not merely absent from output, which is how the
// original bug passed review the first time.
func TestCrawlRefusesOffAllowlistRedirect(t *testing.T) {
	var offsiteHitMu sync.Mutex
	var offsiteHit bool
	offsite := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		offsiteHitMu.Lock()
		offsiteHit = true
		offsiteHitMu.Unlock()
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer offsite.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<a href="/redir">redir</a>`))
	})
	mux.HandleFunc("/redir", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, offsite.URL+"/exfil", http.StatusFound)
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()

	c := NewCrawl("testCrawl")
	c.Url = svr.URL
	c.AllowInsecure = true
	c.Collect()

	offsiteHitMu.Lock()
	hit := offsiteHit
	offsiteHitMu.Unlock()
	assert.False(t, hit, "off-allowlist redirect target must never be contacted")

	assert.Empty(t, c.GetErrors(),
		"a refused redirect on a discovered link is logged, not a fact-level error")
	assert.Empty(t, data.AsMapString(c.GetData()),
		"a refused redirect on a discovered link must not be recorded as a breach")
}

// TestCrawlFollowsSameHostRedirect guards against the allowlist enforcement
// in TestCrawlRefusesOffAllowlistRedirect over-blocking: a same-host
// redirect must still be followed, and links on the final page discovered.
func TestCrawlFollowsSameHostRedirect(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<a href="/hop">hop</a>`))
	})
	mux.HandleFunc("/hop", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/final", http.StatusFound)
	})
	mux.HandleFunc("/final", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<a href="/missing">missing</a>`))
	})
	mux.HandleFunc("/missing", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()

	c := NewCrawl("testCrawl")
	c.Url = svr.URL
	c.AllowInsecure = true
	c.Collect()

	assert.Empty(t, c.GetErrors())
	results := data.AsMapString(c.GetData())
	assert.Equal(t, "404", results[svr.URL+"/missing"],
		"a same-host redirect must still be followed, and its final page's links discovered")
}

// TestCrawlRedirectToIncludeDomainAllowed guards against the allowlist
// check itself, not just its default: a redirect to a host explicitly
// added via include-domains must be followed.
func TestCrawlRedirectToIncludeDomainAllowed(t *testing.T) {
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer other.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<a href="/redir">redir</a>`))
	})
	mux.HandleFunc("/redir", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, other.URL, http.StatusFound)
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()

	c := NewCrawl("testCrawl")
	c.Url = svr.URL
	c.AllowInsecure = true
	c.IncludeDomains = []string{strings.TrimPrefix(other.URL, "http://")}
	c.Collect()

	assert.Empty(t, c.GetErrors())
	results := data.AsMapString(c.GetData())
	assert.Equal(t, "404", results[svr.URL+"/redir"],
		"a redirect to an include-domains host must be followed; the "+
			"pre-redirect url is the recorded key, matching how "+
			"transport-error results are keyed elsewhere in this file")
}

// TestCrawlRedirectLoopBounded is the regression guard for supplying
// CheckRedirect disabling net/http's own default redirect cap: without
// maxCrawlRedirects reimplementing that cap, a self-redirecting endpoint
// would hang the crawl rather than erroring.
func TestCrawlRedirectLoopBounded(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<a href="/loop">loop</a>`))
	})
	mux.HandleFunc("/loop", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/loop", http.StatusFound)
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()

	done := make(chan struct{})
	c := NewCrawl("testCrawl")
	c.Url = svr.URL
	c.AllowInsecure = true
	go func() {
		c.Collect()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("crawl did not terminate - redirect loop was not bounded")
	}

	results := data.AsMapString(c.GetData())
	assert.Contains(t, results[svr.URL+"/loop"], "redirect",
		"a redirect loop must be recorded as an error, not silently dropped")
}

// TestCrawlRootRedirectOffAllowlistIsHardError guards against the crawl
// root itself redirecting off-allowlist: unlike a discovered link (which is
// skipped and logged), this must fail the fact outright. Silently visiting
// nothing and letting not:empty pass would be a worse outcome than an
// explicit, actionable error.
func TestCrawlRootRedirectOffAllowlistIsHardError(t *testing.T) {
	offsite := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer offsite.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, offsite.URL, http.StatusFound)
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()

	c := NewCrawl("testCrawl")
	c.Url = svr.URL
	c.AllowInsecure = true
	c.Collect()

	require.Len(t, c.GetErrors(), 1)
	assert.ErrorIs(t, c.GetErrors()[0], ErrRootRedirectOffAllowlist)
	assert.Nil(t, c.GetData(),
		"a hard error must not also emit a (misleadingly empty, passing) result")
}

func TestCrawlNegativeLimitErrors(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()

	c := NewCrawl("testCrawl")
	c.Url = svr.URL
	c.AllowInsecure = true
	c.Limit = -5
	c.Collect()

	assert.Equal(t, []error{ErrInvalidCrawlLimit}, c.GetErrors())
	assert.Nil(t, c.GetData(),
		"a negative limit must not silently emit an empty, passing result")
}

func TestCrawlNegativeMaxDepthErrors(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()

	c := NewCrawl("testCrawl")
	c.Url = svr.URL
	c.AllowInsecure = true
	c.MaxDepth = -1
	c.Collect()

	assert.Equal(t, []error{ErrInvalidCrawlMaxDepth}, c.GetErrors())
	assert.Nil(t, c.GetData(),
		"a negative max-depth must not silently emit an empty, passing result")
}

// TestCrawlContentTypeCaseInsensitive is the regression guard for Warning
// 2: HTTP media types are case-insensitive (RFC 9110), so a server sending
// an uppercase Content-Type must still have its links discovered.
func TestCrawlContentTypeCaseInsensitive(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "TEXT/HTML")
		w.Write([]byte(`<a href="/missing">missing</a>`))
	})
	mux.HandleFunc("/missing", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()

	c := NewCrawl("testCrawl")
	c.Url = svr.URL
	c.AllowInsecure = true
	c.Collect()

	assert.Empty(t, c.GetErrors())
	results := data.AsMapString(c.GetData())
	assert.Equal(t, "404", results[svr.URL+"/missing"],
		"an uppercase Content-Type must still be parsed for links")
}

func TestCrawlContentTypeXhtml(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xhtml+xml")
		w.Write([]byte(`<a href="/missing">missing</a>`))
	})
	mux.HandleFunc("/missing", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()

	c := NewCrawl("testCrawl")
	c.Url = svr.URL
	c.AllowInsecure = true
	c.Collect()

	assert.Empty(t, c.GetErrors())
	results := data.AsMapString(c.GetData())
	assert.Equal(t, "404", results[svr.URL+"/missing"],
		"application/xhtml+xml must be parsed for links")
}

// TestCrawlNonHtmlContentTypeNotParsed guards against the content-type
// switch in fetchAndExtractLinks being widened too far: a plain-text
// response containing markup-shaped text must not have it parsed for links.
func TestCrawlNonHtmlContentTypeNotParsed(t *testing.T) {
	mux := http.NewServeMux()
	var missingHit bool
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(`<a href="/missing">missing</a>`))
	})
	mux.HandleFunc("/missing", func(w http.ResponseWriter, r *http.Request) {
		missingHit = true
		w.WriteHeader(http.StatusNotFound)
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()

	c := NewCrawl("testCrawl")
	c.Url = svr.URL
	c.AllowInsecure = true
	c.Collect()

	assert.Empty(t, c.GetErrors())
	assert.False(t, missingHit, "text/plain body must not be parsed for links")
	assert.Empty(t, data.AsMapString(c.GetData()))
}

// TestCrawlRootNotFetchedTwice is the regression guard for Warning 4: a
// root url given without a trailing slash ("http://host") and a discovered
// href="/" ("http://host/") resolve to different url.URL.Path values and
// must not be treated as two distinct pages. Asserted via a server-side
// hit counter, since the duplicate is invisible in fact output.
func TestCrawlRootNotFetchedTwice(t *testing.T) {
	var hitsMu sync.Mutex
	hits := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		hitsMu.Lock()
		hits++
		hitsMu.Unlock()
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<a href="/">home</a>`))
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()

	c := NewCrawl("testCrawl")
	c.Url = svr.URL // no trailing slash
	c.AllowInsecure = true
	c.Collect()

	hitsMu.Lock()
	got := hits
	hitsMu.Unlock()
	assert.Equal(t, 1, got,
		"root without a trailing slash and a discovered href=\"/\" must be the same visited-set entry")
	assert.Empty(t, c.GetErrors())
	assert.Empty(t, data.AsMapString(c.GetData()))
}

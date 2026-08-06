package e2e

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCrawlExample exercises the http:crawl fact plugin end-to-end against
// a disposable httptest server, since it needs a live target rather than a
// static fixture tree.
//
// Server link graph:
//
//	/            -> 200, links to /missing, /clean, /no-content
//	/missing     -> 404
//	/no-content  -> 204, anti-regression guard for colly's <203
//	                misclassification bug (must never appear in output)
//	/clean       -> 200, links only to /clean/other
//	/clean/other -> 200, links only back to /clean
//
// The /clean subgraph is closed: from /clean at depth 0, /clean/other is
// depth 1 and links back to /clean which is already visited, so /missing
// (only reachable from /) is never reached regardless of max-depth.
//
// This is also the first e2e coverage of resolve-env: examples/crawl.yml
// sets `url: ${CRAWL_BASE_URL}` with `resolve-env: true`, and this test
// seeds a `.env` file in the staged project directory rather than the OS
// environment, matching pkg/env/envresolver.go's GetEnvMap contract.
func TestCrawlExample(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<a href="/missing">missing</a>` +
			`<a href="/clean">clean</a>` +
			`<a href="/no-content">no content</a>`))
	})
	mux.HandleFunc("/missing", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	mux.HandleFunc("/no-content", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/clean", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<a href="/clean/other">other</a>`))
	})
	mux.HandleFunc("/clean/other", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<a href="/clean">back</a>`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	project := stagedTestdata(t)
	writeFile(t, filepath.Join(project, ".env"), "CRAWL_BASE_URL="+server.URL+"\n")

	res := runShipshape(t, project,
		"run", ".", "-f", examplePath("crawl.yml"), "-o", "json")
	rl := res.DecodeJSON(t)

	siteResult, ok := findResult(rl, "site-has-no-broken-links")
	require.True(t, ok, "site-has-no-broken-links result not present")
	assert.Equal(t, "Fail", siteResult.Status)
	assert.Len(t, siteResult.Breaches, 1,
		"only /missing should breach; /no-content (204) must not")

	cleanResult, ok := findResult(rl, "clean-section-has-no-broken-links")
	require.True(t, ok, "clean-section-has-no-broken-links result not present")
	assert.Equal(t, "Pass", cleanResult.Status)
	assert.Empty(t, cleanResult.Breaches)

	assert.Equal(t, uint32(1), rl.TotalBreaches)

	assert.Equal(t, 0, res.ExitCode, "no -e flag: exit code must be 0 despite the breach")

	// site-has-no-broken-links is severity: high, so -e trips shipshape.Exit's
	// os.Exit(1) path (pkg/shipshape/shipshape.go:206-208).
	resErr := runShipshape(t, project,
		"run", ".", "-f", examplePath("crawl.yml"), "-o", "json", "-e")
	assert.Equal(t, 1, resErr.ExitCode, "-e flag: exit code must be 1 when a high-severity breach is found")
}

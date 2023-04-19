package crawler_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/salsadigitalauorg/shipshape/pkg/checks/crawler"
	"github.com/stretchr/testify/assert"
)

func TestCrawlerMerge(t *testing.T) {
	assert := assert.New(t)

	c := CrawlerCheck{
		Domain:       "foo.example",
		ExtraDomains: []string{"dom1"},
		IncludeURLs:  []string{"url1"},
		Limit:        1,
	}
	c.Merge(&CrawlerCheck{
		Domain:       "bar.example",
		ExtraDomains: []string{"dom2"},
		IncludeURLs:  []string{"url2"},
		Limit:        2,
	})
	assert.EqualValues(CrawlerCheck{
		Domain:       "bar.example",
		ExtraDomains: []string{"dom2"},
		IncludeURLs:  []string{"url2"},
		Limit:        2,
	}, c)
}

func TestCrawlerCheck(t *testing.T) {
	assert := assert.New(t)

	server := httptest.NewServer(
		http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.URL.String() == "/not-found" {
				rw.WriteHeader(http.StatusNotFound)
				rw.Write([]byte(`Not found`))
			} else {
				rw.Write([]byte(`OK`))
			}
		}))
	defer server.Close()

	c := CrawlerCheck{
		IncludeURLs: []string{
			"/not-found",
		},
		Domain: server.URL,
		Limit:  5,
	}

	c.Init(Crawler)
	c.RunCheck()
	assert.ElementsMatch(
		[]string{fmt.Sprintf("Invalid response for: %s/not-found got 404", server.URL)},
		c.Result.Failures,
	)

	c = CrawlerCheck{
		IncludeURLs: []string{},
		Domain:      server.URL,
		Limit:       5,
	}

	c.Init(Crawler)
	c.RunCheck()
	assert.Empty(c.Result.Failures)
	assert.ElementsMatch(
		[]string{"All requests completed successfully"},
		c.Result.Passes,
	)
}

package crawler_test

import (
	"testing"

	. "github.com/salsadigitalauorg/shipshape/pkg/crawler"
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
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

	c := CrawlerCheck{
		IncludeURLs: []string{
			"/not-found",
		},
		Domain: "https://httpbin.org",
		Limit:  5,
	}

	c.Init(shipshape.File)
	c.RunCheck()
	assert.ElementsMatch(
		[]string{"Invalid response for: https://httpbin.org/not-found got 404"},
		c.Result.Failures,
	)

	c = CrawlerCheck{
		IncludeURLs: []string{},
		Domain:      "https://httpbin.org",
		Limit:       5,
	}

	c.Init(shipshape.File)
	c.RunCheck()
	assert.Empty(c.Result.Failures)
	assert.ElementsMatch(
		[]string{"All requests completed successfully"},
		c.Result.Passes,
	)
}

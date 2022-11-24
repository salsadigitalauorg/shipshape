package shipshape_test

import (
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/stretchr/testify/assert"
)

func TestCrawlerMerge(t *testing.T) {
	assert := assert.New(t)

	c := shipshape.CrawlerCheck{
		Domain:       "foo.example",
		ExtraDomains: []string{"dom1"},
		IncludeURLs:  []string{"url1"},
		Limit:        1,
	}
	c.Merge(&shipshape.CrawlerCheck{
		Domain:       "bar.example",
		ExtraDomains: []string{"dom2"},
		IncludeURLs:  []string{"url2"},
		Limit:        2,
	})
	assert.EqualValues(shipshape.CrawlerCheck{
		Domain:       "bar.example",
		ExtraDomains: []string{"dom1", "dom2"},
		IncludeURLs:  []string{"url1", "url2"},
		Limit:        2,
	}, c)
}

func TestCrawlerCheck(t *testing.T) {
	assert := assert.New(t)

	c := shipshape.CrawlerCheck{
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

	c = shipshape.CrawlerCheck{
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

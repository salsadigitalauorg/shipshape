package shipshape_test

import (
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
	"github.com/stretchr/testify/assert"
)

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

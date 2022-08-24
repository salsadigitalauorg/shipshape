package shipshape_test

import (
	"testing"

	"github.com/salsadigitalauorg/shipshape/internal"
	"github.com/salsadigitalauorg/shipshape/pkg/shipshape"
)

func TestCrawlerCheck(t *testing.T) {
	c := shipshape.CrawlerCheck{
		IncludeURLs: []string{
			"/not-found",
		},
		Domain: "https://httpbin.org",
		Limit:  5,
	}

	c.Init(shipshape.File)
	c.RunCheck()

	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string{"Invalid response for: https://httpbin.org/not-found got 404"}); !ok {
		t.Error(msg)
	}

	c = shipshape.CrawlerCheck{
		IncludeURLs: []string{},
		Domain:      "https://httpbin.org",
		Limit:       5,
	}

	c.Init(shipshape.File)
	c.RunCheck()

	if msg, ok := internal.EnsureFailures(t, &c.CheckBase, []string(nil)); !ok {
		t.Error(msg)
	}
	if msg, ok := internal.EnsurePasses(t, &c.CheckBase, []string{"All requests completed successfully"}); !ok {
		t.Error(msg)
	}

}

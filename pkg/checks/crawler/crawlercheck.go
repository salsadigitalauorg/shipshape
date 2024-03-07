package crawler

import (
	"fmt"

	"net/url"

	"github.com/gocolly/colly"
	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/config"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
	"github.com/salsadigitalauorg/shipshape/pkg/utils"
)

//go:generate go run ../../../cmd/gen.go registry --checkpackage=crawler

// CrawlerCheck is a lightweight crawler that can be used to determine
// health of the project.
type CrawlerCheck struct {
	config.CheckBase `yaml:",inline"`
	Domain           string   `yaml:"domain"`
	ExtraDomains     []string `yaml:"extra_domains"`
	IncludeURLs      []string `yaml:"include_urls"`
	Limit            int      `yaml:"limit"`
}

const Crawler config.CheckType = "crawler"

func RegisterChecks() {
	config.ChecksRegistry[Crawler] = func() config.Check { return &CrawlerCheck{} }
}

func init() {
	RegisterChecks()
}

// Merge implementation for file check.
func (c *CrawlerCheck) Merge(mergeCheck config.Check) error {
	crawlerMergeCheck := mergeCheck.(*CrawlerCheck)
	if err := c.CheckBase.Merge(&crawlerMergeCheck.CheckBase); err != nil {
		return err
	}

	utils.MergeString(&c.Domain, crawlerMergeCheck.Domain)
	utils.MergeStringSlice(&c.ExtraDomains, crawlerMergeCheck.ExtraDomains)
	utils.MergeStringSlice(&c.IncludeURLs, crawlerMergeCheck.IncludeURLs)

	if crawlerMergeCheck.Limit > 0 {
		c.Limit = crawlerMergeCheck.Limit
	}
	return nil
}

// RunCheck gathers input configuration and
// prepares the colly crawler to make HTTP requests
// to the project.
//
// @see https://github.com/gocolly/colly/tree/master/_examples
func (c *CrawlerCheck) RunCheck() {
	u, _ := url.Parse(c.Domain)

	allowed_domains := []string{u.Host}
	links := []string{}
	req_count := 0

	allowed_domains = append(allowed_domains, c.ExtraDomains...)

	crawler := colly.NewCollector(
		colly.AllowedDomains(allowed_domains...),
	)

	crawler.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		links = append(links, e.Request.AbsoluteURL(link))
	})

	crawler.OnRequest(func(r *colly.Request) {
		req_count = req_count + 1
	})

	crawler.OnError(func(r *colly.Response, err error) {
		c.Result.Status = result.Fail
		c.AddBreach(&breach.KeyValueBreach{
			Key:        fmt.Sprintf("%v", r.Request.URL),
			ValueLabel: "invalid response",
			Value:      fmt.Sprintf("%d", r.StatusCode),
		})
	})

	crawler.Visit(c.Domain)

	for _, path := range c.IncludeURLs {
		d, _ := url.Parse(c.Domain)
		d.Path = path
		crawler.Visit(d.String())
	}

	for _, link := range links {
		if req_count < c.Limit {
			crawler.Visit(link)
		}
	}

	if c.Result.Status != result.Fail {
		c.Result.Status = result.Pass
		c.AddPass("All requests completed successfully")
	}
}

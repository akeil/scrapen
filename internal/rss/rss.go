package rss

import (
	"context"
	"strings"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"

	"github.com/akeil/scrapen/internal/pipeline"
)

const (
	ctRSS  = "application/rss+xml"
	ctAtom = "application/atom+xml"
	ctXML  = "text/xml"
)

// FindFeeds looks for links to RSS feeds and places them in the `Feeds`
// attribute for the task.
//
// Some deviations from the standard:
// - accept links in <body>
// - does (currently) not de-duplicate links
// - does not require the <base> element to resolve relative URLs
// Source:
// - https://www.rssboard.org/rss-autodiscovery
// - https://developer.mozilla.org/en-US/docs/Web/HTML/Attributes/rel#attr-alternate
//
func FindFeeds(ctx context.Context, t *pipeline.Task) error {
	log.WithFields(log.Fields{
		"task":   t.ID,
		"module": "rss",
	}).Info("Find feeds")

	t.Feeds = make([]pipeline.FeedInfo, 0)
	doc := t.Document()

	doc.Selection.Find("link").Each(func(i int, s *goquery.Selection) {
		rel, _ := s.Attr("rel")
		typ, _ := s.Attr("type")
		href, _ := s.Attr("href")
		title, _ := s.Attr("title")

		rel = strings.ToLower(rel)
		typ = strings.ToLower(typ)

		// exclude non-RSS and incomplete
		if typ != ctRSS && typ != ctAtom && typ != ctXML {
			return
		}
		if rel != "alternate" {
			return
		}
		if href == "" {
			return
		}

		// Found a valid RSS link - make it absolute
		url, err := t.ResolveURL(href)
		if err != nil {
			log.WithFields(log.Fields{
				"task":   t.ID,
				"module": "rss",
				"error":  err,
				"href":   href,
			}).Warning("Failed to resolve feed URL")
			return
		}

		// Found a RSS feed
		log.WithFields(log.Fields{
			"task":   t.ID,
			"module": "rss",
			"rss":    url,
		}).Info("found link")

		t.Feeds = append(t.Feeds, pipeline.FeedInfo{
			URL:   url,
			Title: title,
		})

	})

	return nil
}

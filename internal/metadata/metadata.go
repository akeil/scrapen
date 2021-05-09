package metadata

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"

	"github.com/akeil/scrapen/internal/pipeline"
)

// ReadMetadata reads general metadata from the documents <head>.
func ReadMetadata(ctx context.Context, t *pipeline.Task) error {
	log.WithFields(log.Fields{
		"task":   t.ID,
		"module": "metadata",
		"url":    t.ContentURL(),
	}).Info("Extract metadata")

	doc := t.Document()

	m := newMetadata()
	findMeta(m, doc)
	findLink(m, doc)
	findTitle(m, doc)

	setMetadata(m, t)
	setSite(t)

	return nil
}

func findMeta(m *metadata, doc *goquery.Document) {
	doc.Selection.Find("meta").Each(func(i int, s *goquery.Selection) {
		name, _ := s.Attr("name")
		property, _ := s.Attr("property")
		content, _ := s.Attr("content")

		// property and name should not be present at the same time
		// if they are, prefer name
		if name == "" {
			name = property
		}

		// Do we have any data at all?
		if content == "" || name == "" {
			return
		}

		if contains(descriptionPref, name) {
			m.description[name] = content
			return
		}

		if contains(imagePref, name) {
			m.image[name] = content
			return
		}

		if contains(urlPref, name) {
			m.url[name] = content
			return
		}

		if contains(authorPref, name) {
			m.author[name] = content
			return
		}

		if contains(pubDatePref, name) {
			m.pubDate[name] = content
			return
		}
	})
}

func findLink(m *metadata, doc *goquery.Document) {
	doc.Selection.Find("link").Each(func(i int, s *goquery.Selection) {
		rel, _ := s.Attr("rel")
		href, _ := s.Attr("href")

		if href == "" {
			return
		}

		switch rel {
		case "canonical":
			m.url["link/canonical"] = href
		case "image_src":
			m.image["link/image_src"] = href
		}
	})
}

func findTitle(m *metadata, doc *goquery.Document) {
	doc.Selection.Find("title").First().Each(func(i int, s *goquery.Selection) {
		m.title = strings.TrimSpace(s.Text())
	})
}

var (
	descriptionPref = []string{
		"description",
		"og:description",
		"twitter:description",
		"sailthru.description",
		"preview",
		"krux:description",
	}
	imagePref = []string{
		"og:image:secure_url",
		"og:image:url",
		"og:image",
		"link/image_src",
		"twitter:image",
		"twitter:image:src",
	}
	urlPref = []string{
		"link/canonical",
		"canonicalURL",
		"og:url",
		"twitter:url",
	}
	authorPref = []string{
		"author",
		"article:author",
		"book:author",
		"twitter:creator",
		"parsely-author",
		"sailthru.author",
	}
	pubDatePref = []string{
		"article:published_time",
		"article:modified_time",
		"og:updated_time",
		"date",
		"last-modified",
		"iso-8601-publish-date",
		"parsely-pub-date",
		"sailthru.date",
	}
	// title: og:title, twitter:title, parsely-title, sailthru.title, krux:title
)

func setMetadata(m *metadata, t *pipeline.Task) {
	if m.title != "" {
		t.Title = m.title
	}

	for _, k := range descriptionPref {
		v, ok := m.description[k]
		if ok {
			t.Description = v
			break
		}
	}

	// TODO: resolve the URLK against ContentURL()
	for _, k := range imagePref {
		v, ok := m.image[k]
		if ok {
			t.ImageURL = v
			break
		}
	}

	for _, k := range urlPref {
		v, ok := m.url[k]
		if ok {
			t.CanonicalURL = v
			break
		}
	}

	for _, k := range authorPref {
		v, ok := m.author[k]
		if ok {
			t.Author = v
			break
		}
	}

	for _, k := range pubDatePref {
		v, ok := m.pubDate[k]
		if ok {
			ts := parseTime(v)
			if ts != nil {
				utc := ts.UTC()
				t.PubDate = &utc
			}
			break
		}
	}
}

var prefixes = []string{
	"www.",
	"www1.",
	"www2.",
}

func setSite(t *pipeline.Task) {
	s := t.CanonicalURL
	if s == "" {
		s = t.ContentURL()
	}
	u, err := url.Parse(s)
	if err != nil {
		return
	}
	h := u.Host
	for _, prefix := range prefixes {
		if strings.HasPrefix(h, prefix) {
			h = h[len(prefix):]
		}
	}
	t.Site = h
	t.SiteScheme = u.Scheme
}

type metadata struct {
	title       string
	description map[string]string
	image       map[string]string
	url         map[string]string
	author      map[string]string
	pubDate     map[string]string
}

func newMetadata() *metadata {
	return &metadata{
		description: make(map[string]string),
		image:       make(map[string]string),
		url:         make(map[string]string),
		author:      make(map[string]string),
		pubDate:     make(map[string]string),
	}
}

var layouts = []string{
	time.RFC3339,
	"2006-01-02 15:04:05Z",
}

func parseTime(v string) *time.Time {
	for _, l := range layouts {
		t, err := time.Parse(l, v)
		if err == nil {
			if !t.IsZero() {
				return &t
			}
		}
	}
	return nil
}

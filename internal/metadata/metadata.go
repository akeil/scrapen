package metadata

import (
	"context"
	"net/url"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/akeil/scrapen/internal/pipeline"
)

func ReadMetadata(ctx context.Context, t *pipeline.Task) error {

	log.WithFields(log.Fields{
		"task":   t.ID,
		"module": "metadata",
		"url":    t.ContentURL(),
	}).Info("Extract metadata")

	m := newMetadata()

	reader := func(t html.Token) error {

		tt := t.Type
		switch tt {
		case html.StartTagToken,
			html.SelfClosingTagToken:
			switch t.DataAtom {
			case atom.Meta:
				handleMeta(t, m)
			case atom.Link:
				handleLink(t, m)
			}
		case html.EndTagToken:
		}

		return nil
	}

	err := pipeline.ReadHTML(t.HTML, reader)
	if err != nil {
		return err
	}

	setMetadata(m, t)
	setSite(t)

	return nil
}

var (
	descriptionPref = []string{"description", "og:description", "twitter:description"}
	imagePref       = []string{"og:image:secure_url", "og:image", "link/image_src", "twitter:image", "twitter:image:src"}
	urlPref         = []string{"link/canonical", "og:url", "twitter:url"}
	authorPref      = []string{"author", "article:author", "book:author", "twitter:creator"}
	pubDatePref     = []string{"article:published_time", "article:modified_time", "og:updated_time", "date", "last-modified"}
	// title: og:title
)

func setMetadata(m *metadata, t *pipeline.Task) {
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

func setSite(t *pipeline.Task) {
	u, err := url.Parse(t.ContentURL())
	if err != nil {
		return
	}
	t.Site = u.Host
	t.SiteScheme = u.Scheme
}

type metadata struct {
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

func handleMeta(t html.Token, m *metadata) {
	var name string
	var property string
	var content string
	for _, attr := range t.Attr {
		k := strings.TrimSpace(strings.ToLower(attr.Key))
		v := strings.TrimSpace(attr.Val)
		switch k {
		case "name":
			name = v
		case "property":
			property = v
		case "content":
			content = v
		}
	}

	// property and name should not be present at the same time
	// if they are, prefer name
	if name == "" {
		name = property
	}

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
}

func handleLink(t html.Token, m *metadata) {
	var rel string
	var href string
	for _, attr := range t.Attr {
		k := strings.TrimSpace(strings.ToLower(attr.Key))
		v := strings.TrimSpace(attr.Val)
		switch k {
		case "rel":
			rel = v
		case "href":
			href = v
		}
	}

	if href == "" {
		return
	}

	switch rel {
	case "canonical":
		m.url["link/canonical"] = href
	case "image_src":
		m.image["link/image_src"] = href
	}
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
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

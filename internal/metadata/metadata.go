package metadata

import (
	"context"
	//"fmt"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"akeil.net/akeil/elsewhere/internal/pipeline"
)

func ReadMetadata(ctx context.Context, i *pipeline.Item) (*pipeline.Item, error) {
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

	err := pipeline.ReadHTML(i.Html, reader)
	if err != nil {
		return i, err
	}

	setMetadata(m, i)

	return i, nil
}

var descriptionPref = []string{"description", "og:description", "twitter:description"}
var urlPref = []string{"link/canonical", "og:url", "twitter:url"}

func setMetadata(m *metadata, i *pipeline.Item) {
	for _, k := range descriptionPref {
		v, ok := m.description[k]
		if ok {
			i.Description = v
			break
		}
	}

	for _, k := range urlPref {
		v, ok := m.url[k]
		if ok {
			i.CanonicalURL = v
			break
		}
	}
}

type metadata struct {
	description map[string]string
	url         map[string]string
}

func newMetadata() *metadata {
	return &metadata{
		description: make(map[string]string),
		url:         make(map[string]string),
	}
}

func handleMeta(t html.Token, m *metadata) {
	var name string
	var content string
	for _, attr := range t.Attr {
		k := strings.TrimSpace(strings.ToLower(attr.Key))
		v := strings.TrimSpace(attr.Val)
		switch k {
		case "name":
			name = v
		case "content":
			content = v
		}
	}

	if content == "" {
		return
	}

	if contains(descriptionPref, name) {
		m.description[name] = content
		return
	}

	if contains(urlPref, name) {
		m.url[name] = content
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

	if rel == "canonical" {
		m.url["link/canonical"] = href
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

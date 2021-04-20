package content

import (
	"context"
	"io"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/akeil/scrapen/internal/pipeline"
)

func ResolveURLs(ctx context.Context, t *pipeline.Task) error {

	log.WithFields(log.Fields{
		"task":   t.ID,
		"module": "content",
		"url":    t.ContentURL(),
	}).Info("Resolve URLs in content")

	base, err := url.Parse(t.ContentURL())
	if err != nil {
		return err
	}

	if t.ImageURL != "" {
		img, err := resolve(base, t.ImageURL)
		if err != nil {
			return err
		}
		t.ImageURL = img
	}

	// TODO: replace the std lib parser w/ goquery
	s, err := resolveContentURLs(base, t.HTML())
	if err != nil {
		return err
	}
	t.SetHTML(s)

	return nil
}

func resolveContentURLs(base *url.URL, s string) (string, error) {
	handler := func(tk html.Token, w io.StringWriter) (bool, error) {
		var name string

		switch tk.DataAtom {
		case atom.Img:
			name = "src"
		case atom.A:
			name = "href"
		default:
			return false, nil
		}

		var err error
		tt := tk.Type
		switch tt {
		case html.StartTagToken:
			w.WriteString("<")
			w.WriteString(tk.Data)
			err = resolveAttr(base, tk.Attr, name, w)
			if err != nil {
				return false, err
			}
			w.WriteString(">")
			return true, nil
		case html.SelfClosingTagToken:
			w.WriteString("<")
			w.WriteString(tk.Data)
			err = resolveAttr(base, tk.Attr, name, w)
			if err != nil {
				return false, err
			}
			w.WriteString("/>")
			return true, nil
		}

		return false, nil
	}

	var b strings.Builder
	err := pipeline.WalkHTML(&b, s, handler)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}

func resolveAttr(base *url.URL, a []html.Attribute, name string, w io.StringWriter) error {
	for _, attr := range a {
		if attr.Key == name {
			newHref, err := resolve(base, attr.Val)
			if err != nil {
				return err
			}
			pipeline.WriteAttr(html.Attribute{
				Namespace: "",
				Key:       name,
				Val:       newHref,
			}, w)
		} else {
			pipeline.WriteAttr(attr, w)
		}
	}
	return nil
}

func resolve(base *url.URL, href string) (string, error) {
	ref, err := url.Parse(href)
	if err != nil {
		return "", err
	}

	return base.ResolveReference(ref).String(), nil
}

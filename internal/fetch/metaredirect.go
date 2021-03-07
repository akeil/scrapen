package fetch

import (
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/akeil/scrapen/internal/pipeline"
)

func findRedirect(s string) (string, error) {
	var redirect string

	reader := func(t html.Token) error {
		tt := t.Type
		switch tt {
		case html.StartTagToken,
			html.SelfClosingTagToken:
			switch t.DataAtom {
			case atom.Meta:
				for _, attr := range t.Attr {
					k := strings.ToLower(attr.Key)
					v := strings.ToLower(attr.Val)
					if k == "http-equiv" && v == "refresh" {
						url := parseRefresh(t.Attr)
						// got it
						if url != "" {
							redirect = url
							return nil
						}
					}
				}
			}
		case html.EndTagToken:
		}

		return nil
	}

	err := pipeline.ReadHTML(s, reader)
	if err != nil {
		return "", err
	}

	return redirect, nil
}

func parseRefresh(attrs []html.Attribute) string {
	// expected attribute:
	// content="0;URL=https://buff.ly/3dVxAoz"
	// see:
	// https://developer.mozilla.org/en-US/docs/Web/HTML/Element/meta#attr-http-equiv
	for _, attr := range attrs {
		k := strings.ToLower(attr.Key)
		v := attr.Val
		if k == "content" {
			// refresh can contain
			// - (only) an integer (reload the page) -> we don't care
			// - an Integer followed by `;url=$URL` (redirect to $URL)
			parts := strings.Split(v, ";")
			if len(parts) == 2 {
				// strip the "URL="
				return parts[1][4:]
			}
		}
	}
	return ""
}

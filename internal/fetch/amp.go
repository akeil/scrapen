package fetch

import (
	"errors"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/akeil/scrapen/internal/pipeline"
)

// https://amp.dev/documentation/guides-and-tutorials/learn/spec/amphtml/?referrer=ampproject.org
func findAmpUrl(s string) string {
	var amp string

	reader := func(t html.Token) error {
		tt := t.Type
		switch tt {
		case html.StartTagToken,
			html.SelfClosingTagToken:
			switch t.DataAtom {
			case atom.Link:
				m := make(map[string]string)
				for _, attr := range t.Attr {
					m[strings.ToLower(attr.Key)] = attr.Val
				}

				rel, ok := m["rel"]
				if ok && strings.ToLower(rel) == "amphtml" {
					href, ok := m["href"]
					if ok && href != "" {
						amp = href
						// err is created to stop parsing, will ignore later
						return errors.New("stop parsing")
					}
				}
			}
		case html.EndTagToken:
		}

		return nil
	}

	pipeline.ReadHTML(s, reader)
	return amp
}

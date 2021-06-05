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

const ampScriptPrefix = "https://cdn.ampproject.org/"

// Tell if the document we have fetched is the AMP version of the page.
// If so, swap the content to alternate HTML and alternate URL.
// Return the canonical URL to actual document.
//
// see:
// https://amp.dev/documentation/guides-and-tutorials/learn/spec/amphtml/#required-markup
func checkAMP(s string) (bool, string) {

	var (
		foundHead      bool
		foundBody      bool
		foundCanonical bool
		foundAMPScript bool
		canonical      string
	)

	reader := func(t html.Token) error {
		tt := t.Type
		switch tt {
		case html.StartTagToken,
			html.SelfClosingTagToken:
			switch t.DataAtom {
			case atom.Head:
				foundHead = true
			case atom.Body:
				foundBody = true
			case atom.Link:
				_, rel := pipeline.ReadAttr(t, "rel")
				_, href := pipeline.ReadAttr(t, "href")
				if rel == "canonical" && href != "" {
					foundCanonical = true
					canonical = href
				}
			case atom.Script:
				_, src := pipeline.ReadAttr(t, "src")
				if strings.HasPrefix(src, ampScriptPrefix) {
					foundAMPScript = true
				}
			}
		}

		return nil
	}

	pipeline.ReadHTML(s, reader)

	if foundHead && foundBody && foundCanonical && foundAMPScript {
		return true, canonical
	}

	return false, ""
}

package readable

import (
	"context"
	"strings"

	"github.com/PuerkitoBio/goquery"
	readability "github.com/go-shiori/go-readability"
	log "github.com/sirupsen/logrus"

	"github.com/akeil/scrapen/internal/pipeline"
)

// MakeReadable applies the readability script to each content alternative
// and selects the best content as the HTML for the task.
func MakeReadable(ctx context.Context, t *pipeline.Task) error {
	log.WithFields(log.Fields{
		"task":   t.ID,
		"module": "readable",
		"url":    t.ContentURL(),
	}).Info("Apply readability")

	baseURL := t.ContentURL()
	candidates := make([]candidate, 0)

	a, err := doReadability(t.Document(), baseURL)
	if err != nil {
		return err
	}
	candidates = append(candidates, candidate{baseURL, a})

	altDoc := t.AltDocument()
	if altDoc != nil {
		altA, err := doReadability(altDoc, t.AltURL)
		if err != nil {
			log.WithFields(log.Fields{
				"task":   t.ID,
				"module": "readable",
				"url":    t.ContentURL(),
				"error":  err,
			}).Warning("Readability failed for alternate content")
		} else {
			candidates = append(candidates, candidate{t.AltURL, altA})
		}
	}

	// TODO: keep the alt URL depending on whoch article we selected
	winner := selectArticle(candidates)

	t.SetHTML(winner.Article.Content)
	t.Title = winner.Article.Title
	t.ActualURL = winner.URL

	return nil
}

func doReadability(doc *goquery.Document, baseURL string) (readability.Article, error) {
	var a readability.Article
	p := readability.NewParser()
	s, err := doc.Selection.Find("html").First().Html()
	if err != nil {
		return a, err
	}

	r := strings.NewReader(s)

	a, err = p.Parse(r, baseURL)
	if err != nil {
		return a, err
	}

	return a, nil
}

func selectArticle(candidates []candidate) candidate {
	var result candidate
	maxlen := -1

	for _, c := range candidates {
		r := strings.NewReader(c.Article.Content)
		doc, err := goquery.NewDocumentFromReader(r)
		if err != nil {
			log.WithFields(log.Fields{
				"module": "readable",
				"error":  err,
			}).Warning("Failed to parse Document from content.")
			continue
		}
		// count words
		text := doc.Selection.Find("body").First().Text()
		l := len(text)
		if l > maxlen {
			maxlen = l
			result = c
		}
	}

	log.WithFields(log.Fields{
		"module":       "readable",
		"url":          result.URL,
		"alternatives": len(candidates),
	}).Info("Selected best article by text length")

	return result
}

type candidate struct {
	URL     string
	Article readability.Article
}

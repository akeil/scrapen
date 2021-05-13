package readable

import (
	"context"
	"strings"

	"github.com/PuerkitoBio/goquery"
	readability "github.com/go-shiori/go-readability"
	log "github.com/sirupsen/logrus"

	"github.com/akeil/scrapen/internal/pipeline"
)

func MakeReadable(ctx context.Context, t *pipeline.Task) error {
	log.WithFields(log.Fields{
		"task":   t.ID,
		"module": "readable",
		"url":    t.ContentURL(),
	}).Info("Apply readability")

	baseURL := t.ContentURL()
	candidates := make([]readability.Article, 0)

	a, err := doReadability(t.Document(), baseURL)
	if err != nil {
		return err
	}
	candidates = append(candidates, a)

	altDoc := t.AltDocument()
	if altDoc != nil {
		altA, err := doReadability(altDoc, baseURL)
		if err != nil {
			log.WithFields(log.Fields{
				"task":   t.ID,
				"module": "readable",
				"url":    t.ContentURL(),
				"error":  err,
			}).Warning("Readability failed for alternate content")
		} else {
			candidates = append(candidates, altA)
		}
	}

	article := selectArticle(candidates)

	t.SetHTML(article.Content)
	t.Title = article.Title

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

func selectArticle(candidates []readability.Article) readability.Article {
	var result readability.Article
	maxlen := -1

	for _, a := range candidates {
		r := strings.NewReader(a.Content)
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
			result = a
		}
	}

	return result
}

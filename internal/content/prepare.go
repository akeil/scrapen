package content

import (
	"context"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"

	"github.com/akeil/scrapen/internal/pipeline"
)

// Prepare tries to "fix" the HTML and make it easier to find and extract
// the main content.
func Prepare(ctx context.Context, t *pipeline.Task) error {
	log.WithFields(log.Fields{
		"task":   t.ID,
		"module": "content",
	}).Info("Prepare HTML")

	jsonLD(t)

	doc := t.Document()
	doPrepare(doc)

	altDoc := t.AltDocument()
	if altDoc != nil {
		doPrepare(altDoc)
	}

	return nil
}

func doPrepare(doc *goquery.Document) {
	resolveIFrames(doc)
	resolveNoscriptImage(doc)
	unwrapNoscript(doc)
	unwrapDivs(doc)
	fixSrcs(doc)
	convertAmpImg(doc)
	resolveSrcset(doc)
}

// <noscript> element has a special behaviour in that it is not parsed.
// But it should.
// We want to remove the <noscript> element, but keep its content as HTML.
func unwrapNoscript(doc *goquery.Document) {
	doc.Selection.Find("noscript").Each(func(i int, s *goquery.Selection) {
		// note: do not use noscript.Unwrap()
		// the content of noscript is not parsed as HTML and returns escaped strings
		s.ReplaceWithHtml(s.Text())
	})
}

// if a noscript tag is preceeded by an image, assume that this is the replacement
// image and use it *instead* of the img/picture.
//
// For situations like these:
//
//  <img src="placeholder.jpg" data-lazy-src="actual.jpg" />
//  <noscript>
//    <img src="actual.jpg" />
//  </noscript>
func resolveNoscriptImage(doc *goquery.Document) {
	doc.Find("noscript").Each(func(index int, sel *goquery.Selection) {
		if sel.PrevAll().Length() == 1 {
			siblingName := goquery.NodeName(sel.Prev())
			if siblingName == "picture" || siblingName == "img" {
				sel.Prev().Remove()
				sel.ReplaceWithHtml(sel.Text())
			}
		}
	})

	s, _ := doc.Selection.Find("html").First().Html()
	log.Info(s)
}

// remove div's which wrap a only single element
func unwrapDivs(doc *goquery.Document) {
	doc.Find("div").Each(func(index int, sel *goquery.Selection) {
		if sel.Children().Length() == 1 {
			sel.Unwrap()
		}
	})

	s, _ := doc.Selection.Find("html").First().Html()
	log.Info(s)
}

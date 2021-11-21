package content

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

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

	//log.Debug(t.HTML())

	jsonLD(t)

	doc := t.Document()
	doPrepare(doc)

	altDoc := t.AltDocument()
	if altDoc != nil {
		doPrepare(altDoc)
	}

	//log.Debug(t.HTML())

	return nil
}

func doPrepare(doc *goquery.Document) {
	// Stage 1
	// this may eliminate most of the HTML
	useMain(doc)

	// this makes additional elements visible
	unwrapNoscript(doc)

	// Stage 2
	// dropping elements
	// TODO: *all* of these iterate through the complete doc tree..
	applyRules(rulesPrep, doc)
	dropLinkClouds(doc)
	dropTrackingPixels(doc)

	// Stage 3
	// work on what is left aftr dropping
	resolveIFrames(doc)
	resolveNoscriptImage(doc)
	unwrapDivs(doc)
	dropNavLists(doc)
	fixSrcs(doc)
	convertAmpImg(doc)
	resolveSrcset(doc)
}

func useMain(doc *goquery.Document) {
	hasMain := doc.Find("main").Length() == 1
	if !hasMain {
		return
	}

	main := doc.Find("main").First().Clone()
	doc.Find("html body").First().Children().Remove()
	doc.Find("html body").First().AppendSelection(main)

	log.Info("Replaced content with <main> element")
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
}

// remove div's which wrap a only single element
func unwrapDivs(doc *goquery.Document) {
	doc.Find("div").Each(func(index int, sel *goquery.Selection) {
		if sel.Children().Length() == 1 {
			sel.Unwrap()
		}
	})
}

var whitespace = regexp.MustCompile(`\s`)

// dropNavList attempts to find list elements that are used for navigation
// and removes them.
// "Nav Lists" are lists that have only links as content.
//
// This is done before readability and clean to include links without href
func dropNavLists(doc *goquery.Document) {
	doc.Find("ul, ol").Each(func(i int, s *goquery.Selection) {
		// check if all items consist only of links
		linkOnly := 0
		others := 0
		s.Find("li").Each(func(j int, item *goquery.Selection) {
			a := whitespace.ReplaceAllString(item.Find("a").First().Text(), "")
			b := whitespace.ReplaceAllString(item.Text(), "")
			ratio := float32(len(a)) / float32(len(b))
			if ratio > 0.5 {
				linkOnly++
			} else {
				others++
			}
		})

		if linkOnly >= others {
			log.Debug("Remove list with mostly link-content")
			s.Remove()
		}
	})
}

func dropLinkClouds(doc *goquery.Document) {
	doc.Find("div").Each(func(i int, s *goquery.Selection) {
		a := s.Find("*").Text()
		b := s.Find("a").Text()
		a = whitespace.ReplaceAllString(a, "")
		b = whitespace.ReplaceAllString(b, "")
		aLen := len(a)
		bLen := len(b)

		if bLen == 0 {
			return
		}

		ratio := float32(bLen) / float32(aLen)
		if ratio >= 0.5 {
			log.Debug("Remove link cloud")
			s.Remove()
		}
	})
}

func dropTrackingPixels(doc *goquery.Document) {
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		w, e0 := intAttr("width", s)
		h, e1 := intAttr("height", s)
		// stop here only if *both* dimensions cannot be determined
		if e0 != nil && e1 != nil {
			return
		}

		if w <= 1 || h <= 1 {
			src, _ := s.Attr("src")
			log.WithFields(log.Fields{
				"module": "content",
				"width":  w,
				"height": h,
				"src":    src,
			}).Debug("Remove suspected tracking pixel")
			s.Remove()
		}
	})
}

func intAttr(name string, s *goquery.Selection) (int, error) {
	v, exists := s.Attr(name)
	if !exists {
		return 0, fmt.Errorf("no attribute with name %q", name)
	}

	i, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return 0, err
	}

	return int(i), nil
}

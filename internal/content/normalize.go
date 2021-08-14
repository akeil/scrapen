package content

import (
	"context"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"

	"github.com/akeil/scrapen/internal/pipeline"
)

// Normalize improves the HTML document by (slightly) modifying its content.
func Normalize(ctx context.Context, t *pipeline.Task) error {
	log.WithFields(log.Fields{
		"task":   t.ID,
		"module": "content",
	}).Info("Normalize HTML")

	doc := t.Document()

	fixInlineWhitespace(doc)
	err := normalizeSpace(doc)
	if err != nil {
		return err
	}

	deduplicateImage(t, doc)
	deduplicateTitle(doc, t.Title)
	normalizeHeadings(doc)
	removeMarkupWithinHeadings(doc)

	return nil
}

// attempts to keep spaces *around* inline elements rather than
// *inside them.
//
// foo<em> bar </em>baz  -->  foo <em>bar</em> baz
//        ^   ^                  ^            ^
//
// This is done to cover up the inability of the app's HTML kit to properly
// render whitespace within inline tags.
func fixInlineWhitespace(doc *goquery.Document) {
	doc.Selection.Find("*").Each(func(i int, s *goquery.Selection) {
		tag := goquery.NodeName(s)
		if isInline(tag) {
			t := s.Text()
			prefix := strings.HasPrefix(t, " ")
			suffix := strings.HasSuffix(t, " ")
			if !prefix && !suffix {
				return
			}

			node := s.Get(0)
			if node == nil {
				panic("invalid state") // should never happen
			}

			// precondition, needs to have at least one child node
			if node.FirstChild == nil && node.LastChild != nil {
				return
			}

			// move the space for prefix/suffix out of the inline element
			if prefix && node.FirstChild.Type == html.TextNode {
				if node.PrevSibling != nil && node.PrevSibling.Type == html.TextNode {
					node.PrevSibling.Data += " "
					node.FirstChild.Data = strings.TrimPrefix(node.FirstChild.Data, " ")
				}
			}

			if suffix && node.LastChild.Type == html.TextNode {
				if node.NextSibling != nil && node.NextSibling.Type == html.TextNode {
					node.NextSibling.Data = " " + node.NextSibling.Data
					node.LastChild.Data = strings.TrimSuffix(node.LastChild.Data, " ")
				}
			}

		}
	})
}

// for each blocklevel element, eliminate any whitespace immediately before
// and after the tag.
//
// e.g.:    <h1> Abc </h1>  -->  <h1>Abc</h1>
func normalizeSpace(doc *goquery.Document) error {
	var err error
	doc.Selection.Find("*").Each(func(i int, s *goquery.Selection) {
		tag := goquery.NodeName(s)
		if isBlocklevel(tag) {
			h, e := s.Html()
			if e != nil {
				err = e
				return
			}
			s.SetHtml(strings.TrimSpace(h))
		}
	})

	return err
}

func deduplicateTitle(doc *goquery.Document, title string) {
	doc.Selection.Find("h1, h2, h3, h4, h5, h6").Each(func(i int, s *goquery.Selection) {
		t := strings.TrimSpace(s.Text())
		if strings.EqualFold(t, title) {
			s.Remove()
		}
	})
}

// If the same image URL that is the "main" image for the article
// also appears in the content, remove it from content.
func deduplicateImage(t *pipeline.Task, doc *goquery.Document) {
	if t.ImageURL == "" {
		return
	}

	doc.Selection.Find("img").Each(func(i int, s *goquery.Selection) {
		src, exists := s.Attr("src")
		if !exists {
			return
		}

		if src == t.ImageURL {
			s.Remove()
		}
	})
}

var headings = []string{"h1", "h2", "h3", "h4", "h5", "h6"}

// normalizeHeadings reorders heading levels (h1 through h6). The result is a
// consistent structure of headings without gaps between the levels.
// Normalized headings start with h1.
func normalizeHeadings(doc *goquery.Document) {
	count := map[string]int{}

	// collect the number of occurences for each heading level
	for _, k := range headings {
		doc.Selection.Find(k).Each(func(i int, s *goquery.Selection) {
			count[k]++
		})
	}

	// move all headers up by "n" levels, so that the highest level reaches h1
	// fill gaps between headings
	//var target string
	target := -1
	for i, k := range headings {

		if count[k] == 0 {
			if target < 0 {
				target = i
			}
		} else {
			if target >= 0 {
				doc.Selection.Find(k).Each(func(i int, s *goquery.Selection) {
					html, _ := s.Html()
					h := headings[target]
					new := fmt.Sprintf("<%v>%v</%v>", h, html, h)
					s.ReplaceWithHtml(new)
				})
				target++
			}
		}
	}
}

// Remove markup within headings
func removeMarkupWithinHeadings(doc *goquery.Document) {
	match := strings.Join(headings, ",")

	doc.Find(match).Each(func(i int, s *goquery.Selection) {
		s.Children().Each(func(i int, child *goquery.Selection) {
			child.Contents().Unwrap()
		})
	})
}

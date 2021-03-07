package content

import (
	"context"
	"strings"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"

	"github.com/akeil/scrapen/internal/pipeline"
)

func Clean(ctx context.Context, t *pipeline.Task) error {

	log.WithFields(log.Fields{
		"task":   t.ID,
		"module": "content",
	}).Info("Clean HTML")

	r := strings.NewReader(t.HTML)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return err
	}

	// TODO: does not really belong here
	resolvePicture(doc)

	removeUnwantedElements(doc)
	unwrapTags(doc)
	removeUnwantedAttributes(doc)

	html, err := doc.Selection.Find("body").First().Html()
	if err != nil {
		return err
	}
	t.HTML = html

	return nil
}

// dropUnwantedTags finds tags from the greylists an removes the *markup*
// but keeps the text content
func unwrapTags(doc *goquery.Document) {
	doc.Selection.Find("*").Each(func(i int, s *goquery.Selection) {
		tag := goquery.NodeName(s)
		if shouldUnwrap(tag) {
			// cannot unwrap empty
			if s.Contents().Length() == 0 {
				s.Remove()
			} else {
				s.Contents().Unwrap()
			}
		}
	})

	unwrapAnchors(doc)
}

func unwrapAnchors(doc *goquery.Document) {
	doc.Selection.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists || href == "" {
			s.Contents().Unwrap()
		}
	})
}

// contains HTML elements which we do not want in the output.
// But we DO want to keep their text content.
var unwrap = []string{
	"span", "div",
	"article", "section", "summary",
	"address",
	"main", "footer", "header", "nav",
	"hgroup",
	"data",
	"dfn",
	// deprecated elements
	"acronym", "basefont", "big", "blink", "center",
	"content", "font", "listing",
	"marquee", "nobr", "plaintext", "spacer",
	"strike", "tt",
	"picture",
}

func shouldUnwrap(tag string) bool {
	for _, greylisted := range unwrap {
		if tag == greylisted {
			return true
		}
	}

	return false
}

// remove unwanted attribs
func removeUnwantedAttributes(doc *goquery.Document) {
	doc.Selection.Find("*").Each(func(i int, s *goquery.Selection) {
		tag := goquery.NodeName(s)
		// needs to be done in two steps,
		// editing Node.Attribs is not allowed while iterating
		remove := make([]string, 0)
		for _, attr := range s.Nodes[0].Attr {
			if !isWhitelistedAttr(tag, attr.Key) {
				remove = append(remove, attr.Key)
			}
		}
		for _, key := range remove {
			s.RemoveAttr(key)
		}
	})
}

// keep `id` for all elements?
// to support internal links?
var attrWhitelist = map[string][]string{
	"img": []string{"src", "width", "height", "alt", "title"},
	"a":   []string{"href", "title"}, // rel?
	//"svg":   []string{"xmlns", "viewBox", "version", "x", "y", "style"},
	//"path":   []string{"d"},
}

func isWhitelistedAttr(tag, attr string) bool {
	whitelist, ok := attrWhitelist[tag]
	if !ok {
		return false
	}

	for _, whitelisted := range whitelist {
		if attr == whitelisted {
			return true
		}
	}

	return false
}

// remove any non-whitelisted tags - including their content/children
func removeUnwantedElements(doc *goquery.Document) {
	doc.Selection.Find("*").Each(func(i int, s *goquery.Selection) {
		tag := goquery.NodeName(s)
		if !isWhitelistedTag(tag) {
			s.Remove()
		}
	})
}

// whitelist contains HTML elements which we want to keep.
var whitelist = []string{
	"#document", "html", "body",
	"p",
	"a",
	"h1", "h2", "h3", "h4", "h5", "h6",
	"br", "hr",
	"b", "u", "i", "s",
	"em", "strong", "small",
	"sub", "sup",
	"abbr",
	"del", "ins",
	"aside",
	"ul", "ol", "li",
	"dl", "dd", "dt",
	"table", "thead", "tbody", "tfoot", "caption", "tr", "th", "td", "colgroup", "col",
	"code", "pre", "kbd", "sample", "var",
	"mark", "q",
	"rp", "rt", "rtc", "ruby",
	"blockquote", "cite",
	"img",
	"figure", "figcaption",
	"bdi", "bdo",
	"time",
	"wbr",
	// audio, video, track, source
	// embed, iframe,
	// object, param,
	// picture, source
	// svg, path, g
}

func isWhitelistedTag(tag string) bool {
	for _, whitelisted := range whitelist {
		if tag == whitelisted {
			return true
		}
	}
	// tags we want empty, but NOT removed
	return shouldUnwrap(tag)
}
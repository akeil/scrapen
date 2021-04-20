package content

import (
	"context"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"

	"github.com/akeil/scrapen/internal/pipeline"
)

func Clean(ctx context.Context, t *pipeline.Task) error {

	log.WithFields(log.Fields{
		"task":   t.ID,
		"module": "content",
	}).Info("Clean HTML")

	doc := t.Document()

	// TODO: does not really belong here
	resolvePicture(doc)

	removeUnwantedElements(doc)
	unwrapTags(doc)
	removeUnwantedAttributes(doc)

	return nil
}

// dropUnwantedTags finds tags from the greylists an removes the *markup*
// but keeps the text content
func unwrapTags(doc *goquery.Document) {
	doc.Selection.Find("*").Each(func(i int, s *goquery.Selection) {
		tag := goquery.NodeName(s)
		if isGraylisted(tag) {
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

// remove any non-whitelisted tags - including their content/children
func removeUnwantedElements(doc *goquery.Document) {
	doc.Selection.Find("*").Each(func(i int, s *goquery.Selection) {
		tag := goquery.NodeName(s)
		if !isWhitelistedTag(tag) {
			s.Remove()
		}
	})
}

package specific

import (
	"fmt"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
	//"golang.org/x/net/html"
	//"golang.org/x/net/html/atom"

	"github.com/akeil/scrapen/internal/pipeline"
)

func stackoverflow(t *pipeline.Task) {
	log.WithFields(log.Fields{
		"task":   t.ID,
		"src":    t.ContentURL(),
		"module": "content",
	}).Debug("Apply stackoverflow")

	doc := t.Document()
	c := doc.Clone()

	q := c.Find(".question .js-post-body").First()
	if q.Length() == 0 {
		// question not found, leave content unchanged.
		return
	}

	// replace the full content with an empty article and the question
	doc.Find("*").First().SetHtml("<article></article>")
	doc.Find("article").First().AppendHtml("<h2>Question</h2>")
	doc.Find("article").First().AppendSelection(q)

	// add the accepted answer
	// itemprop=acceptedAnswer
	answerCount := 0
	q = c.Find(".accepted-answer .js-post-body").First()
	if q.Length() > 0 {
		doc.Find("article").First().AppendHtml("<hr />")
		doc.Find("article").First().AppendHtml("<h2>Accepted Answer</h2>")
		doc.Find("article").First().AppendSelection(q)
		answerCount++
	}

	// add other answers
	// itemprop=sugesstedAnswer
	c.Find(".answer .js-post-body").Each(func(index int, sel *goquery.Selection) {
		n := answerCount + 1 + index
		doc.Find("article").First().AppendHtml("<hr />")
		doc.Find("article").First().AppendHtml(fmt.Sprintf("<h2>Answer #%v</h2>", n))
		doc.Find("article").First().AppendSelection(sel)
	})
}

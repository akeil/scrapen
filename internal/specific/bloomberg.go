package specific

import (
	"encoding/json"
	"strings"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"

	"github.com/akeil/scrapen/internal/pipeline"
)

// bloomberg implements spcific handling for bloomberg.com
func bloomberg(t *pipeline.Task) {
	log.WithFields(log.Fields{
		"task":   t.ID,
		"src":    t.ContentURL(),
		"module": "specific",
	}).Debug("Apply bloomberg")

	doc := t.Document()

	// check if the article content is delivered inside a <script> tag
	// as JSON:
	//
	//  <script type="application/json" data-component-props="ArticleBody">
	//	{
	//      "body": "<div class=\"inline-newsletter-top\"> ...",
	//      "teaserBody": "...",
	//      "marketCards": "",
	//      "bottomLeftTouts": "...",
	//      "id": "Q9RBPQT1UM1301",
	//      "fortressEnabled": true,
	//      "shouldLoadDataDash": false
	//  }
	//  </script>

	doc.Find("script").EachWithBreak(func(i int, s *goquery.Selection) bool {
		typ, ok := s.Attr("type")
		if !ok || typ != "application/json" {
			return true
		}

		dcp, ok := s.Attr("data-component-props")
		if !ok || dcp != "ArticleBody" {
			return true
		}

		// Found article JSON in <script>

		r := strings.NewReader(s.Text())
		dec := json.NewDecoder(r)
		data := make(map[string]interface{})
		err := dec.Decode(&data)
		if err != nil {
			log.WithFields(log.Fields{
				"task":   t.ID,
				"src":    t.ContentURL(),
				"err":    err,
				"module": "specific",
			}).Debug("Failed to parse JSON-Article-Data")
		}

		body, ok := data["body"]
		if !ok {
			return false
		}

		// replace all content with the new body
		html := body.(string)
		doc.Find("*").First().SetHtml("<article></article>")
		doc.Find("article").First().AppendHtml(html)

		log.WithFields(log.Fields{
			"task":   t.ID,
			"src":    t.ContentURL(),
			"module": "specific",
		}).Debug("Replaced content")

		return false
	})
}

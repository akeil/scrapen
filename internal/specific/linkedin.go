package specific

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"

	"github.com/akeil/scrapen/internal/pipeline"
)

const (
	// this is a *partial* part of the class for a shared article or post div
	linkedinArticleClass = "share-article"
	linkedinPostClass    = "share-update-card"
)

func linkedin(t *pipeline.Task) {
	log.WithFields(log.Fields{
		"task":   t.ID,
		"src":    t.ContentURL(),
		"module": "content",
	}).Debug("Apply linkedin")

	// if you share a post/an article on linkedin,
	// thesharing link points to a linkedin landing page.
	// This landing page contains the link to the actual article.
	//
	// landing page URLs look like this:
	//
	// linkedin.com/posts/<IDENTIFIER>
	//
	// The Markup around the link looks like this:
	//
	//   <div class="feed-shared-article__link-container">
	//     <a class="app-aware-link feed-shared-article__image-link tap-target"
	//        href="https://buff.ly/IDENTIFIER" />
	//   <div class="feed-shared-article__description-container">
	//     <div class="flex-grow-1">
	//       <a class="app-aware-link feed-shared-article__meta flex-grow-1 full-width tap-target"
	//          href="https://buff.ly/IDENTIFIER" />
	//
	// IF a "normal" post is shared, we will see a div with

	doc := t.Document()
	isPost := false
	isArticle := false

	doc.Find("div").Each(func(index int, s *goquery.Selection) {
		c, _ := s.Attr("class")
		classes := strings.Split(c, " ")
		log.Debug(classes)
		for _, class := range classes {
			class = strings.TrimSpace(class)
			log.Debug(class)

			if strings.HasPrefix(class, linkedinArticleClass) {
				log.Info("Found shared article")
				isArticle = true
			} else if strings.HasPrefix(class, linkedinPostClass) {
				log.Info("Found shared post")
				isPost = true
			}
		}
	})

	if isPost {
		sharedLinkedinPost(t)
		return
	}
}

func sharedLinkedinPost(t *pipeline.Task) {
	doc := t.Document()
	c := doc.Clone()

	post := c.Find("div.share-update-card").First()
	doc.Find("*").First().SetHtml("<article></article>")

	// we need to add a title - use the poster's name
	// a class="share-update-card__actor-text-link"
	var title string
	post.Find("header").Find("a").Each(func(index int, sel *goquery.Selection) {
		if title == "" {
			title = strings.TrimSpace(sel.Text())
		}
	})

	if title != "" {
		t.Title = title
		doc.Find("article").First().AppendHtml("<h1>" + html.EscapeString(title) + "</h1>")
		doc.Find("head").First().AppendHtml("<title>" + html.EscapeString(title) + "</title>")
	}

	doc.Find("article").First().AppendSelection(post)
}

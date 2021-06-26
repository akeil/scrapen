package content

import (
	"context"
	"strings"

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
	dropBlacklisted(doc)
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
}

// remove div's which wrap a only single element
func unwrapDivs(doc *goquery.Document) {
	doc.Find("div").Each(func(index int, sel *goquery.Selection) {
		if sel.Children().Length() == 1 {
			sel.Unwrap()
		}
	})
}

// see AMP components list:
// https://amp.dev/documentation/components/?format=websites
var blacklist = []string{
	"amp-ad",
	"amp-ad-exit",
	"amp-analytics",
	"amp-auto-ads",
	"amp-call-tracking",
	"amp-social-share",
	"amp-sticky-ad",

	// almos all of the "dynamic content" elements
	"amp-access-laterpay",
	"amp-access-poool",
	"amp-access-scroll",
	"amp-access",
	"amp-action-macro",
	"amp-autocomplete",
	"amp-bind",
	"amp-byside-content",
	"amp-consent",
	"amp-date-picker",
	"amp-form",
	"amp-geo",
	"amp-gist",
	"amp-google-assistant-assistjs",
	"amp-google-document-embed",
	"amp-inpumask",
	"amp-install-serviceworker",
	//"amp-layout",
	"amp-link-rewriter",
	"amp-list",
	"amp-live-list",
	"amp-minute-media-player",
	"amp-mustache",
	"amp-next-page",
	"amp-pixel",
	"amp-recaptcha-input",
	"amp-render",
	"amp-script",
	"amp-selector",
	"amp-smartlinks",
	"amp-subscriptions-google",
	"amp-subscriptions",
	"amp-user-notification",
	"amp-video-docking",
	"amp-web-push",

	// social
	"amp-addthis",
	"amp-beopinion",
	"amp-facebook-comments",
	"amp-facebook-like",
	"amp-facebook-page",
	"amp-facebook",
	"amp-gfycat",
	"amp-iframely",
	"amp-instagram",
	"amp-pinterest",
	"amp-reddit",
	"amp-riddle-quiz",
	"amp-twitter",
	"amp-vine",
	"amp-vk",

	// social login
	"amp-onetap-google",

	"amp-state",

	"template",
}

// Drop all unwantedelements including their content.
// This is used to get rid of everything that is easier to detect *before*
// readability is applied.
func dropBlacklisted(doc *goquery.Document) {
	match := strings.Join(blacklist, ",")
	doc.Find(match).Each(func(index int, sel *goquery.Selection) {
		sel.Remove()
	})
}

package content

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
)

func TestUnwrap(t *testing.T) {
	assert := assert.New(t)

	d := doc("<span>foo</span>")
	unwrapTags(d)
	assert.Equal("foo", str(d))

	d = doc("prefix <span>text</span> suffix")
	unwrapTags(d)
	assert.Equal("prefix text suffix", str(d))

	d = doc("prefix <span></span> suffix")
	unwrapTags(d)
	assert.Equal("prefix  suffix", str(d))

	d = doc("<div>prefix <span>text</span> suffix</div>")
	unwrapTags(d)
	assert.Equal("prefix text suffix", str(d))
}

func TestEmptyAnchor(t *testing.T) {
	assert := assert.New(t)

	d := doc("<a href=\"\">text</a>")
	unwrapAnchors(d)
	assert.Equal("text", str(d))

	d = doc("<a>text</a>")
	unwrapAnchors(d)
	assert.Equal("text", str(d))
}

func TestRemoveAttributes(t *testing.T) {
	assert := assert.New(t)

	d := doc("<p>text</p>")
	removeUnwantedAttributes(d)
	assert.Equal("<p>text</p>", str(d))

	d = doc("<p style=\"something\">text</p>")
	removeUnwantedAttributes(d)
	assert.Equal("<p>text</p>", str(d))

	d = doc("<img src=\"foo\" style=\"bar\"/>")
	removeUnwantedAttributes(d)
	assert.Equal("<img src=\"foo\"/>", str(d))
}

func TestRemoveElements(t *testing.T) {
	assert := assert.New(t)

	d := doc("<p>text<iframe><p>content</p></iframe></p>")
	removeUnwantedElements(d)
	assert.Equal("<p>text</p>", str(d))

	d = doc("text<script></script>text")
	removeUnwantedElements(d)
	assert.Equal("texttext", str(d))
}

func doc(s string) *goquery.Document {
	r := strings.NewReader(s)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		panic(err)
	}
	return doc
}

func str(doc *goquery.Document) string {
	html, err := doc.Selection.Find("body").First().Html()
	if err != nil {
		panic(err)
	}
	return html
}

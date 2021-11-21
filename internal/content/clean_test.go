package content

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"

	"github.com/akeil/scrapen/internal/pipeline"
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

func TestRemovePunctuation(t *testing.T) {
	assert := assert.New(t)

	d := doc("<p>text</p><p>|</p>")
	removeUnwantedPunctuation(d)
	assert.Equal("<p>text</p>", str(d))

	d = doc("<p>text</p><p> | </p>")
	removeUnwantedPunctuation(d)
	assert.Equal("<p>text</p>", str(d))

	d = doc("<p>| text |</p>")
	removeUnwantedPunctuation(d)
	assert.Equal("<p>| text |</p>", str(d))
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

func TestDropEmptyElements(t *testing.T) {
	assert := assert.New(t)

	d := doc("<p>text</p><p></p><p>text</p>")
	dropEmptyElements(d)
	assert.Equal("<p>text</p><p>text</p>", str(d))

	// nested
	d = doc("<p>text</p><p><strong></strong></p><p>text</p>")
	dropEmptyElements(d)
	assert.Equal("<p>text</p><p>text</p>", str(d))

	// empty allowed
	d = doc("<p>text<br/>text</p>")
	dropEmptyElements(d)
	assert.Equal("<p>text<br/>text</p>", str(d))
}

func TestDropChildlessParents(t *testing.T) {
	assert := assert.New(t)

	d := doc("text<ol></ol>text")
	dropChildlessParents(d)
	assert.Equal("texttext", str(d))

	// (invalid) text content is also dropped
	d = doc("text<ol>INVALID</ol>text")
	dropChildlessParents(d)
	assert.Equal("texttext", str(d))

	// keep if valid
	d = doc("<ol><li>foo</li></ol>")
	dropChildlessParents(d)
	assert.Equal("<ol><li>foo</li></ol>", str(d))

	// not affected
	d = doc("<em>no children</em>")
	dropChildlessParents(d)
	assert.Equal("<em>no children</em>", str(d))
}

func TestRemoveUnsupportedScheme(t *testing.T) {
	assert := assert.New(t)

	d := doc("<img src=\"http://foo.png\"/>")
	removeUnsupportedSchemes(d)
	assert.Equal("<img src=\"http://foo.png\"/>", str(d))

	d = doc("<img src=\"https://foo.png\"/>")
	removeUnsupportedSchemes(d)
	assert.Equal("<img src=\"https://foo.png\"/>", str(d))

	d = doc("<img src=\"data:BASE64\"/>")
	removeUnsupportedSchemes(d)
	assert.Equal("<img src=\"data:BASE64\"/>", str(d))

	d = doc("<img src=\"\"/>")
	removeUnsupportedSchemes(d)
	assert.Equal("<img src=\"\"/>", str(d))

	// we need this to work as long as we resolve URLs *after* clean
	d = doc("<img src=\"image.jpg\"/>")
	removeUnsupportedSchemes(d)
	assert.Equal("<img src=\"image.jpg\"/>", str(d))

	d = doc("<p>unchanged</p>")
	removeUnsupportedSchemes(d)
	assert.Equal("<p>unchanged</p>", str(d))
}

func TestNormalizeUrls(t *testing.T) {
	assert := assert.New(t)

	d := doc(`<img src="normal.png"/>`)
	normalizeUrls(d)
	assert.Equal(`<img src="normal.png"/>`, str(d))

	d = doc(`<img src="https://normal.png"/>`)
	normalizeUrls(d)
	assert.Equal(`<img src="https://normal.png"/>`, str(d))

	// e.g. (URL) _> URL
	d = doc(`<img src="(https://normal.png)"/>`)
	normalizeUrls(d)
	assert.Equal(`<img src="https://normal.png"/>`, str(d))

	// or " URL " -> "URL"
	d = doc(`<img src=" https://normal.png "/>`)
	normalizeUrls(d)
	assert.Equal(`<img src="https://normal.png"/>`, str(d))

	// "something URL" -> "URL"
	d = doc(`<img src="head https://normal.png"/>`)
	normalizeUrls(d)
	assert.Equal(`<img src="https://normal.png"/>`, str(d))

	// "something URL" -> "URL"
	d = doc(`<img src="https://normal.png tail"/>`)
	normalizeUrls(d)
	assert.Equal(`<img src="https://normal.png"/>`, str(d))
}

func TestStripFromTitle(t *testing.T) {
	assert := assert.New(t)
	task := &pipeline.Task{}

	task.Title = "No Prefix"
	stripFromTitle(task)
	assert.Equal("No Prefix", task.Title)

	task.Title = "Prefix | The Actual Title"
	stripFromTitle(task)
	assert.Equal("The Actual Title", task.Title)

	task.Title = "The Actual Title | Suffix"
	stripFromTitle(task)
	assert.Equal("The Actual Title", task.Title)

	// with site name as prefix or suffix w/ various separators
	task.Title = "The Actual Title - The Site Name"
	task.SiteName = "The Site Name"
	stripFromTitle(task)
	assert.Equal("The Actual Title", task.Title)

	task.Title = "The Actual Title: The Site Name"
	task.SiteName = "The Site Name"
	stripFromTitle(task)
	assert.Equal("The Actual Title", task.Title)

	task.Title = "The Site Name | The Actual Title"
	task.SiteName = "The Site Name"
	stripFromTitle(task)
	assert.Equal("The Actual Title", task.Title)

	task.Title = "The Site Name: The Actual Title"
	task.SiteName = "The Site Name"
	stripFromTitle(task)
	assert.Equal("The Actual Title", task.Title)

	// not to be stripped
	task.Title = "The Actual Title can contain The Site Name"
	task.SiteName = "The Site Name"
	stripFromTitle(task)
	assert.Equal("The Actual Title can contain The Site Name", task.Title)
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

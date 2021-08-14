package content

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/akeil/scrapen/internal/pipeline"
)

func TestNormalizeSpace(t *testing.T) {
	assert := assert.New(t)

	d := doc("<p>unchanged</p>")
	normalizeSpace(d)
	assert.Equal("<p>unchanged</p>", str(d))

	d = doc("<p> with space </p>")
	normalizeSpace(d)
	assert.Equal("<p>with space</p>", str(d))

	d = doc("<p> with <em>some</em> space </p>")
	normalizeSpace(d)
	assert.Equal("<p>with <em>some</em> space</p>", str(d))

	d = doc("<p> with<a> some </a>space </p>")
	normalizeSpace(d)
	assert.Equal("<p>with<a> some </a>space</p>", str(d))
}

func TestFixInineWhitespace(t *testing.T) {
	assert := assert.New(t)

	d := doc("<p>foo<em> bar </em>baz</p>")
	fixInlineWhitespace(d)
	assert.Equal("<p>foo <em>bar</em> baz</p>", str(d))

	// prefix only
	d = doc("<p>foo<em> bar</em> baz</p>")
	fixInlineWhitespace(d)
	assert.Equal("<p>foo <em>bar</em> baz</p>", str(d))

	// suffix only
	d = doc("<p>foo <em>bar </em>baz</p>")
	fixInlineWhitespace(d)
	assert.Equal("<p>foo <em>bar</em> baz</p>", str(d))

	// no previous sibling w. text
	d = doc("<p><em> bar </em>baz</p>")
	fixInlineWhitespace(d)
	assert.Equal("<p><em> bar</em> baz</p>", str(d))

	// empty element not supported
	d = doc("<p>foo<em><strong> bar </strong></em>baz</p>")
	fixInlineWhitespace(d)
	assert.Equal("<p>foo<em><strong> bar </strong></em>baz</p>", str(d))

	// nested nodes
	d = doc("<p>foo<em></em>baz</p>")
	fixInlineWhitespace(d)
	assert.Equal("<p>foo<em></em>baz</p>", str(d))
}

func TestDeduplicateTitle(t *testing.T) {
	assert := assert.New(t)

	d := doc("<h2>My Title</h2><p>unchanged</p>")
	deduplicateTitle(d, "My Title")
	assert.Equal("<p>unchanged</p>", str(d))

	// case insensitive
	d = doc("<h2>MY TITLE</h2><p>unchanged</p>")
	deduplicateTitle(d, "My Title")
	assert.Equal("<p>unchanged</p>", str(d))

	// keep other headings
	d = doc("<h2>My Title</h2><p>unchanged</p> <h3>Other Title</h3>")
	deduplicateTitle(d, "My Title")
	assert.Equal("<p>unchanged</p> <h3>Other Title</h3>", str(d))
}

func TestDeduplicateImage(t *testing.T) {
	assert := assert.New(t)

	// no images anywhere - should have no effect
	task := &pipeline.Task{}
	d := doc("<p>Some text</p>")
	deduplicateImage(task, d)
	assert.Equal("", task.ImageURL)
	assert.Equal("<p>Some text</p>", str(d))

	// no image URL - should have no effect
	d = doc("<p>Some <img src=\"image.jpg\"/>text</p>")
	deduplicateImage(task, d)
	assert.Equal("", task.ImageURL)
	assert.Equal("<p>Some <img src=\"image.jpg\"/>text</p>", str(d))

	// same URL in image and content - should drop from content
	d = doc("<p>Some <img src=\"image.jpg\"/>text</p>")
	task.ImageURL = "image.jpg"
	deduplicateImage(task, d)
	assert.Equal("image.jpg", task.ImageURL)
	assert.Equal("<p>Some text</p>", str(d))

	// same URL in image and TWICE content - should drop from content
	d = doc("<p>Some <img src=\"image.jpg\"/>text<img src=\"image.jpg\"/></p>")
	task.ImageURL = "image.jpg"
	deduplicateImage(task, d)
	assert.Equal("image.jpg", task.ImageURL)
	assert.Equal("<p>Some text</p>", str(d))

	// should not remove different URLs
	d = doc("<p>Some <img src=\"image.jpg\"/>text</p>")
	task.ImageURL = "different.jpg"
	deduplicateImage(task, d)
	assert.Equal("different.jpg", task.ImageURL)
	assert.Equal("<p>Some <img src=\"image.jpg\"/>text</p>", str(d))
}

func TestNormalizeHeadings(t *testing.T) {
	assert := assert.New(t)

	// leave content w/o headings unchanged
	d := doc("<p>Some text</p>")
	normalizeHeadings(d)
	assert.Equal("<p>Some text</p>", str(d))

	// leave consistent headings alone
	d = doc("<h1>Some text</h1><p>foo</p><h2>Heading 2</h2>")
	normalizeHeadings(d)
	assert.Equal("<h1>Some text</h1><p>foo</p><h2>Heading 2</h2>", str(d))

	// close gaps (1, 3 -> 1, 2)
	d = doc("<h1>Some text</h1><p>foo</p><h3>Heading 2</h3>")
	normalizeHeadings(d)
	assert.Equal("<h1>Some text</h1><p>foo</p><h2>Heading 2</h2>", str(d))

	// close multiple gaps (1, 3, 6 -> 1, 2, 3)
	d = doc("<h1>Some text</h1><p>foo</p><h3>Heading 2</h3><p>foo</p><h6>Heading 6</h6>")
	normalizeHeadings(d)
	assert.Equal("<h1>Some text</h1><p>foo</p><h2>Heading 2</h2><p>foo</p><h3>Heading 6</h3>", str(d))

	// move up to h1
	d = doc("<h2>Some text</h2><p>foo</p><h3>Heading 2</h3>")
	normalizeHeadings(d)
	assert.Equal("<h1>Some text</h1><p>foo</p><h2>Heading 2</h2>", str(d))

	// move up and close gaps
	d = doc("<h2>Some text</h2><p>foo</p><h5>Heading 2</h5>")
	normalizeHeadings(d)
	assert.Equal("<h1>Some text</h1><p>foo</p><h2>Heading 2</h2>", str(d))
}

func TestUnwrapWithinHeadings(t *testing.T) {
	assert := assert.New(t)

	// leave content w/o headings unchanged
	d := doc("<h1><p>Foo</p> Bar</h1>")
	removeMarkupWithinHeadings(d)
	assert.Equal("<h1>Foo Bar</h1>", str(d))

	d = doc("<h1>Foo <small>Bar</small></h1>")
	removeMarkupWithinHeadings(d)
	assert.Equal("<h1>Foo Bar</h1>", str(d))

	d = doc("<h1>Foo Bar</h1>")
	removeMarkupWithinHeadings(d)
	assert.Equal("<h1>Foo Bar</h1>", str(d))
}

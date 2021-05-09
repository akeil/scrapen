package content

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/akeil/scrapen/internal/pipeline"
)

func TestResolveURLs(t *testing.T) {
	assert := assert.New(t)

	var task *pipeline.Task

	task = &pipeline.Task{
		ActualURL: "https://example.com/abc/index.html",
		ImageURL:  "/images/image.jpg",
	}
	task.SetHTML("")

	err := ResolveURLs(nil, task)
	assert.Nil(err)
	assert.Equal("https://example.com/images/image.jpg", task.ImageURL)

	// resolve relative
	resolveTest(t, task, `<a href="./relative.html">foo</a>`, `<a href="https://example.com/abc/relative.html">foo</a>`)

	// resolve absolute
	resolveTest(t, task, `<a href="/absolute.html">foo</a>`, `<a href="https://example.com/absolute.html">foo</a>`)

	// external links unchanged
	resolveTest(t, task, `<a href="https://elsewhere.com/index.html">foo</a>`, `<a href="https://elsewhere.com/index.html">foo</a>`)

	// img
	resolveTest(t, task, `<img src="/images/img.jpg"/>`, `<img src="https://example.com/images/img.jpg"/>`)
}

func resolveTest(t *testing.T, task *pipeline.Task, html, expected string) {
	assert := assert.New(t)

	task.SetHTML(html)
	err := ResolveURLs(nil, task)
	assert.Nil(err)

	// goquery will automatically insert head and body
	expected = "<head></head><body>" + expected + "</body>"
	assert.Equal(expected, task.HTML())
}

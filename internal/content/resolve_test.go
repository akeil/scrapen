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

	err := ResolveURLs(nil, task)
	assert.Nil(err)
	assert.Equal("https://example.com/images/image.jpg", task.ImageURL)

	// resolve relative
	task.SetHTML(`<a href="./relative.html">foo</a>`)
	err = ResolveURLs(nil, task)
	assert.Nil(err)
	assert.Equal(`<a href="https://example.com/abc/relative.html">foo</a>`, task.HTML())

	// resolve absolute
	task.SetHTML(`<a href="/absolute.html">foo</a>`)
	err = ResolveURLs(nil, task)
	assert.Nil(err)
	assert.Equal(`<a href="https://example.com/absolute.html">foo</a>`, task.HTML())

	// external links unchanged
	task.SetHTML(`<a href="https://elsewhere.com/index.html">foo</a>`)
	err = ResolveURLs(nil, task)
	assert.Nil(err)
	assert.Equal(`<a href="https://elsewhere.com/index.html">foo</a>`, task.HTML())

	// img
	task.SetHTML(`<img src="/images/img.jpg"/>`)
	err = ResolveURLs(nil, task)
	assert.Nil(err)
	assert.Equal(`<img src="https://example.com/images/img.jpg"/>`, task.HTML())
}

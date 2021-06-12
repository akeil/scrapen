package content

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/akeil/scrapen/internal/pipeline"
)

func TestResolveURLs(t *testing.T) {
	assert := assert.New(t)

	task := &pipeline.Task{
		ActualURL: "https://example.com/abc/index.html",
		ImageURL:  "/images/image.jpg",
	}
	task.SetHTML("")

	err := ResolveURLs(context.TODO(), task)
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

	// tilde
	resolveTest(t, task, `<img src="/data/images/name~_img.jpg"/>`, `<img src="https://example.com/data/images/name~_img.jpg"/>`)
}

func TestTildeURL(t *testing.T) {
	assert := assert.New(t)

	html := `<picture>
		<source media="(max-width: 420px)" data-srcset="/multimedia/bilder/alex-barck-101~_v-mittelgross1x1.jpg" />
		<source media="(min-width: 1024px)" data-srcset="/multimedia/bilder/alex-barck-101~_v-videowebl.jpg" />
		<img src="placeholder.jpg" />
	</picture>`
	d := doc(html)
	resolvePicture(d)

	img := d.Selection.Find("img").First()
	src, _ := img.Attr("src")
	assert.Equal("/multimedia/bilder/alex-barck-101~_v-videowebl.jpg", src)
}

func resolveTest(t *testing.T, task *pipeline.Task, html, expected string) {
	assert := assert.New(t)

	task.SetHTML(html)
	err := ResolveURLs(context.TODO(), task)
	assert.Nil(err)

	// goquery will automatically insert head and body
	expected = "<head></head><body>" + expected + "</body>"
	assert.Equal(expected, task.HTML())
}

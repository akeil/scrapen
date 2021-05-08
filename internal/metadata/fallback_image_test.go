package metadata

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/akeil/scrapen/internal/pipeline"
)

func TestFallbackImage(t *testing.T) {
	assert := assert.New(t)

	var err error
	task := &pipeline.Task{}

	// image already set
	task.ImageURL = "something.jpg"
	err = FallbackImage(nil, task)
	assert.Nil(err)
	assert.Equal("something.jpg", task.ImageURL)

	// no image in metadata, use first from content
	html := `<html><head></head><body>
		<p>Foo</p>
		<img src="image.jpg"/>
		<p>Bar</p>
		<img src="other.jpg"/>
	</body></html>`

	task.ImageURL = ""
	task.SetHTML(html)

	err = FallbackImage(nil, task)
	assert.Nil(err)
	assert.Equal("image.jpg", task.ImageURL)

	// from Favicon, by size
	html = `<html><head>
        <link rel="icon" type="image/png" sizes="32x32" href="small.png">
        <link rel="icon" type="image/png" sizes="100x100" href="icon.png">
    </head><body><p>Content</p></body></html>
    `
	task.SetHTML(html)
	task.ImageURL = ""
	err = FallbackImage(nil, task)
	assert.Nil(err)
	assert.Equal("icon.png", task.ImageURL)

	// from Favicon, by preference
	html = `<html><head>
        <link rel="apple-touch-icon" type="image/png" href="icon.png">
        <link rel="icon" type="image/png" href="pref.png">
    </head><body><p>Content</p></body></html>
    `
	task.SetHTML(html)
	task.ImageURL = ""
	err = FallbackImage(nil, task)
	assert.Nil(err)
	assert.Equal("pref.png", task.ImageURL)
}

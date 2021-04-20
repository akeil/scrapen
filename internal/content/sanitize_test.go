package content

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/akeil/scrapen/internal/pipeline"
)

func TestSanitize(t *testing.T) {
	// accept stuuf
	callSanitize(t, "", "")
	callSanitize(t, "<p>foo</p>", "<p>foo</p>")
	callSanitize(t, "<img src=\"myimage.jpg\"/>", "<img src=\"myimage.jpg\"/>")
	callSanitize(t, "<a href=\"example.com\">text</a>", "<a href=\"example.com\">text</a>")

	// remove stuff
	callSanitize(t, "<a onclick=\"alert('')\">text</a>", "text")
	callSanitize(t, "<a onclick=\"alert('')\" href\"example.com\">text</a>", "text")
	callSanitize(t, "<script>foo()</script>", "")
	callSanitize(t, "<style>abc</style>", "")
	callSanitize(t, "<iframe>abc</iframe>", "")

	callSanitize(t, "<video>alternate</video>", "alternate")

	// partial remove
	callSanitize(t, "abc <script>foo()</script> def", "abc  def")
}

func callSanitize(t *testing.T, html, expected string) {
	assert := assert.New(t)
	task := &pipeline.Task{}
	task.SetHTML(html)
	ctx := context.TODO()
	var err error

	err = Sanitize(ctx, task)
	assert.Nil(err)
	assert.Equal(expected, task.HTML())
}

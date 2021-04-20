package assets

import (
	"errors"
	"testing"

	"github.com/akeil/scrapen/internal/pipeline"
	"github.com/stretchr/testify/assert"
)

func TestImg(t *testing.T) {
	html := `<img src="https://example.com/image.jpg"/>`
	expect := `<img src="store://ID"/>`
	doFetchImages(t, html, expect)
}

func TestFigure(t *testing.T) {
	html := `<figure><img src="https://example.com/image.jpg"/></figure>`
	expect := `<figure><img src="store://ID"/></figure>`
	doFetchImages(t, html, expect)
}

func TestEmpty(t *testing.T) {
	doFetchImages(t, "", "")
}

func TestPlain(t *testing.T) {
	doFetchImages(t, "abc", "abc")
}

func TestBasicHTML(t *testing.T) {
	html := "<p>abc</p>"
	doFetchImages(t, html, html)
}

func TestHTMLAttributes(t *testing.T) {
	html := `<p class="foo" id="bar"><span class="baz">abc</span></p>`
	doFetchImages(t, html, html)
}

func TestSelfClosingTag(t *testing.T) {
	html := `foo<br/>baz`
	doFetchImages(t, html, html)
}

func TestFetchError(t *testing.T) {
	assert := assert.New(t)
	task := pipeline.Task{
		URL: "https://example.com/base",
	}
	task.SetHTML(`<img src="https://example.com/image.jpg"/>`)
	fetch := func(s string) (string, error) {
		return "", errors.New("test error")
	}
	err := doImages(fetch, &task)
	assert.Nil(err)
	assert.Equal(`<head></head><body><img src="https://example.com/image.jpg"/></body>`, task.HTML())
}

func doFetchImages(t *testing.T, html, expected string) {
	assert := assert.New(t)

	i := pipeline.Task{
		URL: "https://example.com/base",
	}
	i.SetHTML(html)

	fetch := func(s string) (string, error) {
		return "store://ID", nil
	}

	err := doImages(fetch, &i)

	expected = "<head></head><body>" + expected + "</body>"

	assert.Nil(err)
	assert.Equal(expected, i.HTML())
}

func TestDataURL(t *testing.T) {
	assert := assert.New(t)
	task := pipeline.NewTask(pipeline.NewMemoryStore(), "task-id", "https://example.com")
	task.SetHTML(`<html><body>
		<p>Text</p>
		<img src="data:image/jpeg;base64,SGVsbG8sIFdvcmxkIQ=="/>
	</body></html>`)

	err := DownloadImages(nil, task)
	assert.Nil(err)
	assert.Equal(1, len(task.Images))
	assert.NotEqual("", task.Images[0].ContentURL)
	assert.Equal("image/jpeg", task.Images[0].ContentType)
}

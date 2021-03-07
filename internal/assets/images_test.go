package assets

import (
	"errors"
	"testing"

	"github.com/akeil/scrapen/internal/pipeline"
	"github.com/stretchr/testify/assert"
)

func TestImg(t *testing.T) {
	assert := assert.New(t)
	html := `<img src="https://example.com/image.jpg"/>`
	expect := `<img src="store://ID"/>`
	i, err := doFetchImages(html)
	assert.Nil(err)
	assert.Equal(expect, i.HTML)
}

func TestFigure(t *testing.T) {
	assert := assert.New(t)
	html := `<figure><img src="https://example.com/image.jpg"/></figure>`
	expect := `<figure><img src="store://ID"/></figure>`
	i, err := doFetchImages(html)
	assert.Nil(err)
	assert.Equal(expect, i.HTML)
}

func TestEmpty(t *testing.T) {
	assert := assert.New(t)
	html := ""
	i, err := doFetchImages(html)
	assert.Nil(err)
	assert.Equal(html, i.HTML)
}

func TestPlain(t *testing.T) {
	assert := assert.New(t)
	html := "abc"
	i, err := doFetchImages(html)
	assert.Nil(err)
	assert.Equal(html, i.HTML)
}

func TestBasicHTML(t *testing.T) {
	assert := assert.New(t)
	html := "<p>abc</p>"
	i, err := doFetchImages(html)
	assert.Nil(err)
	assert.Equal(html, i.HTML)
}

func TestHTMLAttributes(t *testing.T) {
	assert := assert.New(t)
	html := `<p class="foo" id="bar"><span class="baz">abc</span></p>`
	i, err := doFetchImages(html)
	assert.Nil(err)
	assert.Equal(html, i.HTML)
}

func TestSelfClosingTag(t *testing.T) {
	assert := assert.New(t)
	html := `foo<br/>baz`
	i, err := doFetchImages(html)
	assert.Nil(err)
	assert.Equal(html, i.HTML)
}

func doFetchImages(html string) (pipeline.Task, error) {
	i := pipeline.Task{
		URL:  "https://example.com/base",
		HTML: html,
	}

	fetch := func(s string) (string, error) {
		return "ID", nil
	}

	return i, doImages(fetch, &i)
}

func TestFetchError(t *testing.T) {
	assert := assert.New(t)
	task := pipeline.Task{
		URL:  "https://example.com/base",
		HTML: `<img src="https://example.com/image.jpg"/>`,
	}
	fetch := func(s string) (string, error) {
		return "", errors.New("test error")
	}
	err := doImages(fetch, &task)
	assert.Nil(err)
	assert.Equal(`<img src="https://example.com/image.jpg"/>`, task.HTML)
}

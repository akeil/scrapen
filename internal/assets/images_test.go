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
	expect := `<img src="local://ID"/>`
	i, err := doFetchImages(html)
	assert.Nil(err)
	assert.Equal(expect, i.Html)
}

func TestFigure(t *testing.T) {
	assert := assert.New(t)
	html := `<figure><img src="https://example.com/image.jpg"/></figure>`
	expect := `<figure><img src="local://ID"/></figure>`
	i, err := doFetchImages(html)
	assert.Nil(err)
	assert.Equal(expect, i.Html)
}

func TestEmpty(t *testing.T) {
	assert := assert.New(t)
	html := ""
	i, err := doFetchImages(html)
	assert.Nil(err)
	assert.Equal(html, i.Html)
}

func TestPlain(t *testing.T) {
	assert := assert.New(t)
	html := "abc"
	i, err := doFetchImages(html)
	assert.Nil(err)
	assert.Equal(html, i.Html)
}

func TestBasicHTML(t *testing.T) {
	assert := assert.New(t)
	html := "<p>abc</p>"
	i, err := doFetchImages(html)
	assert.Nil(err)
	assert.Equal(html, i.Html)
}

func TestHTMLAttributes(t *testing.T) {
	assert := assert.New(t)
	html := `<p class="foo" id="bar"><span class="baz">abc</span></p>`
	i, err := doFetchImages(html)
	assert.Nil(err)
	assert.Equal(html, i.Html)
}

func TestSelfClosingTag(t *testing.T) {
	assert := assert.New(t)
	html := `foo<br/>baz`
	i, err := doFetchImages(html)
	assert.Nil(err)
	assert.Equal(html, i.Html)
}

func doFetchImages(html string) (pipeline.Item, error) {
	i := pipeline.Item{
		Url:  "https://example.com/base",
		Html: html,
	}

	fetch := func(s string) (string, error) {
		return "ID", nil
	}

	return i, doImages(fetch, &i)
}

func TestFetchError(t *testing.T) {
	assert := assert.New(t)
	i := pipeline.Item{
		Url:  "https://example.com/base",
		Html: `<img src="https://example.com/image.jpg"/>`,
	}
	fetch := func(s string) (string, error) {
		return "", errors.New("test error")
	}
	err := doImages(fetch, &i)
	assert.NotNil(err)
}

func TestResolve(t *testing.T) {
	assert := assert.New(t)
	var s string
	var err error
	s, err = resolveURL("image.jpg", "https://example.com")
	assert.Nil(err)
	assert.Equal("https://example.com/image.jpg", s)

	s, err = resolveURL("assets/image.jpg", "https://example.com/path/index.html")
	assert.Nil(err)
	assert.Equal("https://example.com/path/assets/image.jpg", s)

	s, err = resolveURL("/assets/image.jpg", "https://example.com/path/index.html")
	assert.Nil(err)
	assert.Equal("https://example.com/assets/image.jpg", s)

	s, err = resolveURL("assets/image.jpg", "https://example.com/path/index.html#id")
	assert.Nil(err)
	assert.Equal("https://example.com/path/assets/image.jpg", s)

	s, err = resolveURL("https://cdn.org/image.jpg", "https://example.com")
	assert.Nil(err)
	assert.Equal("https://cdn.org/image.jpg", s)
}

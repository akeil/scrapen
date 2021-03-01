package metadata

import (
	//"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/akeil/scrapen/internal/pipeline"
)

func TestMeta(t *testing.T) {
	assert := assert.New(t)
	html := `<html><head>
        <meta name="foo" content="bar" />
    </head><body>foo</body></html>`
	i := &pipeline.Item{
		HTML: html,
	}

	_, err := ReadMetadata(nil, i)
	assert.Nil(err)
}

func TestMetaDescription(t *testing.T) {
	assert := assert.New(t)

	// basic
	html := `<html><head>
        <meta name="description" content="the description" />
    </head><body>foo</body></html>`

	i, err := readMeta(html)
	assert.Nil(err)
	assert.Equal("the description", i.Description)

	// og:
	html = `<html><head>
        <meta property="og:description" content="the description" />
    </head><body>foo</body></html>`

	i, err = readMeta(html)
	assert.Nil(err)
	assert.Equal("the description", i.Description)

	// twitter:
	html = `<html><head>
        <meta property="twitter:description" content="the description" />
    </head><body>foo</body></html>`

	i, err = readMeta(html)
	assert.Nil(err)
	assert.Equal("the description", i.Description)

	// preference
	html = `<html><head>
        <meta name="description" content="the description" />
        <meta property="og:description" content="NOT the description" />
    </head><body>foo</body></html>`

	i, err = readMeta(html)
	assert.Nil(err)
	assert.Equal("the description", i.Description)

	// unsupported
	html = `<html><head>
        <meta name="xxx:description" content="the description" />
    </head><body>foo</body></html>`

	i, err = readMeta(html)
	assert.Nil(err)
	assert.Equal("", i.Description)
}

func TestImageURL(t *testing.T) {
	assert := assert.New(t)

	// og
	html := `<html><head>
        <meta property="og:image" content="https://example.com/foo.jpg" />
    </head><body>foo</body></html>`

	i, err := readMeta(html)
	assert.Nil(err)
	assert.Equal("https://example.com/foo.jpg", i.ImageURL)

	// link
	html = `<html><head>
        <link rel="image_src" href="https://example.com/foo.jpg" />
    </head><body>foo</body></html>`

	i, err = readMeta(html)
	assert.Nil(err)
	assert.Equal("https://example.com/foo.jpg", i.ImageURL)

	// preference
	html = `<html><head>
        <meta property="og:image" content="https://example.com/foo.jpg" />
        <meta property="twitter:image" content="https://example.com/IGNORED.jpg" />
    </head><body>foo</body></html>`

	i, err = readMeta(html)
	assert.Nil(err)
	assert.Equal("https://example.com/foo.jpg", i.ImageURL)
}

func TestCanonicalURL(t *testing.T) {
	assert := assert.New(t)

	// og
	html := `<html><head>
        <meta property="og:url" content="https://example.com/foo" />
    </head><body>foo</body></html>`

	i, err := readMeta(html)
	assert.Nil(err)
	assert.Equal("https://example.com/foo", i.CanonicalURL)

	// link
	html = `<html><head>
        <link rel="canonical" href="https://example.com/foo" />
    </head><body>foo</body></html>`

	i, err = readMeta(html)
	assert.Nil(err)
	assert.Equal("https://example.com/foo", i.CanonicalURL)

	// preference
	html = `<html><head>
        <meta property="og:url" content="https://example.com/IGNORE" />
        <link rel="canonical" href="https://example.com/foo" />
    </head><body>foo</body></html>`

	i, err = readMeta(html)
	assert.Nil(err)
	assert.Equal("https://example.com/foo", i.CanonicalURL)

}

func readMeta(html string) (*pipeline.Item, error) {
	i := &pipeline.Item{
		HTML: html,
	}

	return ReadMetadata(nil, i)
}

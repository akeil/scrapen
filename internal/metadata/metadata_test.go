package metadata

import (
	//"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"akeil.net/akeil/elsewhere/internal/pipeline"
)

func TestMeta(t *testing.T) {
	assert := assert.New(t)
	html := `<html><head>
        <meta name="foo" content="bar" />
    </head><body>foo</body></html>`
	i := &pipeline.Item{
		Html: html,
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
        <meta name="og:description" content="the description" />
    </head><body>foo</body></html>`

	i, err = readMeta(html)
	assert.Nil(err)
	assert.Equal("the description", i.Description)

	// twitter:
	html = `<html><head>
        <meta name="twitter:description" content="the description" />
    </head><body>foo</body></html>`

	i, err = readMeta(html)
	assert.Nil(err)
	assert.Equal("the description", i.Description)

	// preference
	html = `<html><head>
        <meta name="description" content="the description" />
        <meta name="og:description" content="NOT the description" />
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

func TestCanonicalURL(t *testing.T) {
	assert := assert.New(t)

	// og
	html := `<html><head>
        <meta name="og:url" content="https://example.com/foo" />
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
        <meta name="og:url" content="https://example.com/IGNORE" />
        <link rel="canonical" href="https://example.com/foo" />
    </head><body>foo</body></html>`

	i, err = readMeta(html)
	assert.Nil(err)
	assert.Equal("https://example.com/foo", i.CanonicalURL)

}

func readMeta(html string) (*pipeline.Item, error) {
	i := &pipeline.Item{
		Html: html,
	}

	return ReadMetadata(nil, i)
}

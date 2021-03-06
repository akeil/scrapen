package fetch

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindRedirect(t *testing.T) {
	assert := assert.New(t)

	html := `<head>
        <meta http-equiv="refresh" content="0;URL=https://example.com">
    </head>`
	s, err := findRedirect(html)
	assert.Nil(err)
	assert.Equal("https://example.com", s)

	// UPPERCASE tag name
	html = `<head>
        <META http-equiv="refresh" content="0;URL=https://example.com">
    </head>`
	s, err = findRedirect(html)
	assert.Nil(err)
	assert.Equal("https://example.com", s)

	// lowercase URL param
	html = `<head>
        <meta http-equiv="refresh" content="0;url=https://example.com">
    </head>`
	s, err = findRedirect(html)
	assert.Nil(err)
	assert.Equal("https://example.com", s)

	// no redirect
	html = `<head>
        <meta http-equiv="refresh" content="123">
    </head>`
	s, err = findRedirect(html)
	assert.Nil(err)
	assert.Equal("", s)

	// no meta
	html = `<head>
        <meta foo="bar" content="123">
    </head>`
	s, err = findRedirect(html)
	assert.Nil(err)
	assert.Equal("", s)
}

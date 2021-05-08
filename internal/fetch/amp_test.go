package fetch

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindAMP(t *testing.T) {
	assert := assert.New(t)

	html := `<head>
        <link rel="amphtml" href="https://example.com/amp.html">
    </head>`
	s := findAmpUrl(html)
	assert.Equal("https://example.com/amp.html", s)

	// No amp
	html = `<head>
        <link rel="canonical" href="https://example.com/doc.html">
    </head>`
	s = findAmpUrl(html)
	assert.Equal("", s)

	// Supports UPPERCASE for rel, href is case-sensitive
	html = `<head>
        <LINK REL="AMPHTML" HREF="https://example.com/AMP.html">
    </head>`
	s = findAmpUrl(html)
	assert.Equal("https://example.com/AMP.html", s)

	// use the first entry
	html = `<head>
        <link rel="amphtml" href="https://example.com/amp.html">
        <link rel="amphtml" href="https://example.com/duplicate.html">
    </head>`
	s = findAmpUrl(html)
	assert.Equal("https://example.com/amp.html", s)
}

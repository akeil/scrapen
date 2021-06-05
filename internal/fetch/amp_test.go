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

func TestCheckAMP(t *testing.T) {
	assert := assert.New(t)

	// all criteria are met
	html := `<html>
	<head>
        <link rel="canonical" href="https://example.com/canonical.html">
		<script async="async" src="https://cdn.ampproject.org/v0.js"></script>
    </head>
	<body>
		<p>Content</p>
	</body>
	<html>`
	isAMP, canonical := checkAMP(html)
	assert.True(isAMP)
	assert.Equal("https://example.com/canonical.html", canonical)

	// The canonical link alone is not enough
	html = `<head>
        <link rel="canonical" href="https://example.com/canonical.html">
    </head>`
	isAMP, _ = checkAMP(html)
	assert.False(isAMP)

	// AMP script alone is not enough
	html = `<head>
        <script async="async" src="https://cdn.ampproject.org/v0.js"></script>
    </head>`
	isAMP, _ = checkAMP(html)
	assert.False(isAMP)
}

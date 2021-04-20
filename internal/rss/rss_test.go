package rss

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/akeil/scrapen/internal/pipeline"
)

func TestFindFeeds(t *testing.T) {
	assert := assert.New(t)

	base := "https://example.com"
	html := `<html><head>
        <link rel="alternate" type="application/rss+xml" title="-&gt; RSS" href="rss.xml"/>
    </head><body>foo</body></html>`

	fi, err := findRss(base, html)
	assert.Nil(err)
	assert.NotNil(fi)
	assert.Equal(1, len(fi))
	if len(fi) == 1 {
		i := fi[0]
		assert.Equal("-> RSS", i.Title)
		assert.Equal("https://example.com/rss.xml", i.URL)
	}

	// UPPERCASE
	html = `<html><head>
        <LINK REL="ALTERNATE" TYPE="application/rss+xml" TITLE="RSS" HREF="rss.xml"/>
    </head><body>foo</body></html>`

	fi, err = findRss(base, html)
	assert.Nil(err)
	assert.NotNil(fi)
	assert.Equal(1, len(fi))

	// different URL
	html = `<html><head>
        <link rel="alternate" type="application/rss+xml" href="https://feed.example.com/rss.xml"/>
    </head><body>foo</body></html>`

	fi, err = findRss(base, html)
	assert.Nil(err)
	assert.Equal(1, len(fi))
	if len(fi) == 1 {
		i := fi[0]
		assert.Equal("https://feed.example.com/rss.xml", i.URL)
	}

	// not an RSS link
	html = `<html><head>
        <link rel="stylesheet" type="text/css" href="style.css"/>
    </head><body>foo</body></html>`

	fi, err = findRss(base, html)
	assert.Nil(err)
	assert.Equal(0, len(fi))

	// invalid RSS link, missing rel
	html = `<html><head>
        <link type="application/rss+xml" href="rss.xml"/>
    </head><body>foo</body></html>`

	fi, err = findRss(base, html)
	assert.Nil(err)
	assert.Equal(0, len(fi))

	// invalid RSS link, missing type
	html = `<html><head>
        <link rel="alternate" href="rss.xml"/>
    </head><body>foo</body></html>`

	fi, err = findRss(base, html)
	assert.Nil(err)
	assert.Equal(0, len(fi))

	// invalid RSS link, missing href
	html = `<html><head>
        <link rel="alternate" type="application/rss+xml" rel="alternate" href=""/>
    </head><body>foo</body></html>`

	fi, err = findRss(base, html)
	assert.Nil(err)
	assert.Equal(1, len(fi))

	// invalid RSS link, missing type` AND rel
	html = `<html><head>
        <link href="rss.xml"/>
    </head><body>foo</body></html>`

	fi, err = findRss(base, html)
	assert.Nil(err)
	assert.Equal(0, len(fi))

	// Multiple feeds
	html = `<html><head>
        <link rel="alternate" type="application/rss+xml" title="Foo" href="foo.xml" />
        <link rel="alternate" type="application/rss+xml" title="Bar" href="bar.xml" />
        <link rel="alternate" type="application/rss+xml" title="Baz" href="baz.xml" />
    </head><body>foo</body></html>`

	fi, err = findRss(base, html)
	assert.Nil(err)
	assert.NotNil(fi)
	assert.Equal(3, len(fi))
	if len(fi) == 3 {
		assert.Equal("Foo", fi[0].Title)
		assert.Equal("https://example.com/foo.xml", fi[0].URL)

		assert.Equal("Bar", fi[1].Title)
		assert.Equal("https://example.com/bar.xml", fi[1].URL)

		assert.Equal("Baz", fi[2].Title)
		assert.Equal("https://example.com/baz.xml", fi[2].URL)
	}
}

func findRss(base, html string) ([]pipeline.FeedInfo, error) {
	task := &pipeline.Task{
		ActualURL: base,
	}
	task.SetHTML(html)

	err := FindFeeds(nil, task)
	return task.Feeds, err
}

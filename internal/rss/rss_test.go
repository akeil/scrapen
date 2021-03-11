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
        <link rel="alternate" type="application/rss+xml" title="RSS" href="rss.xml"/>
    </head><body>foo</body></html>`

	fi, err := findRss(base, html)
	assert.Nil(err)
	assert.NotNil(fi)
	assert.Equal(1, len(fi))
	if len(fi) == 1 {
		i := fi[0]
		assert.Equal("RSS", i.Title)
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

}

func findRss(base, html string) ([]pipeline.FeedInfo, error) {

	task := &pipeline.Task{
		HTML:      html,
		ActualURL: base,
	}

	err := FindFeeds(nil, task)
	return task.FeedInfo, err
}

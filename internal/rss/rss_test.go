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
        <link rel="alternate" type="application/rss+xml" title="RSS" href="../rss.xml"/>
    </head><body>foo</body></html>`

	fi, err := findRss(base, html)
	assert.Nil(err)
	assert.NotNil(fi)
	assert.Equal(1, len(fi))

	// UPPERCASE
	html = `<html><head>
        <LINK REL="ALTERNATE" TYPE="application/rss+xml" TITLE="RSS" HREF="../rss.xml"/>
    </head><body>foo</body></html>`

	fi, err = findRss(base, html)
	assert.Nil(err)
	assert.NotNil(fi)
	assert.Equal(1, len(fi))

}

func findRss(base, html string) ([]pipeline.FeedInfo, error) {

	task := &pipeline.Task{
		HTML:      html,
		ActualURL: base,
	}

	err := FindFeeds(nil, task)
	return task.FeedInfo, err
}

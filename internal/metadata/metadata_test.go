package metadata

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/akeil/scrapen/internal/pipeline"
)

func TestMeta(t *testing.T) {
	assert := assert.New(t)
	html := `<html><head>
        <meta name="foo" content="bar" />
    </head><body>foo</body></html>`
	i := &pipeline.Task{}
	i.SetHTML(html)

	err := ReadMetadata(nil, i)
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

func TestUnescape(t *testing.T) {
	assert := assert.New(t)

	// basic
	html := `<html><head>
		<title>the &quot;title&quot;</title>
        <meta name="description" content="the &quot;description&quot;" />
    </head><body>foo</body></html>`

	i, err := readMeta(html)
	assert.Nil(err)
	assert.Equal("the \"title\"", i.Title)
	assert.Equal("the \"description\"", i.Description)
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

func TestAuthor(t *testing.T) {
	assert := assert.New(t)

	// basic
	html := `<html><head>
        <meta property="author" content="The Author" />
    </head><body>foo</body></html>`

	i, err := readMeta(html)
	assert.Nil(err)
	assert.Equal("The Author", i.Author)

	// preference
	html = `<html><head>
		<meta property="twitter:creator" content="NOT The Author" />
        <meta property="article:author" content="The Author" />
    </head><body>foo</body></html>`

	i, err = readMeta(html)
	assert.Nil(err)
	assert.Equal("The Author", i.Author)

}

func TestSite(t *testing.T) {
	assert := assert.New(t)

	i := &pipeline.Task{
		URL: "https://foo.bar.com/path?query#fragment",
	}
	setSite(i)
	assert.Equal("foo.bar.com", i.Site)
	assert.Equal("https", i.SiteScheme)

	i = &pipeline.Task{
		URL: "http://foo.bar.com/path?query#fragment",
	}
	setSite(i)
	assert.Equal("foo.bar.com", i.Site)
	assert.Equal("http", i.SiteScheme)

	// Preference - ActualURL before CanonicalURL
	i.URL = "https://shorten.it/xyz"
	i.ActualURL = "https://mobile.foo.com/path/page.html"
	i.CanonicalURL = "https://foo.com/path/page.html"
	setSite(i)
	assert.Equal("foo.com", i.Site)

	i.URL = "https://shorten.it/xyz"
	i.ActualURL = ""
	i.CanonicalURL = "https://foo.com/path/page.html"
	setSite(i)
	assert.Equal("foo.com", i.Site)

	// strip www. prefix
	i.URL = "https://www.example.com/xyz"
	i.ActualURL = ""
	i.CanonicalURL = ""
	setSite(i)
	assert.Equal("example.com", i.Site)
}

func TestParseTime(t *testing.T) {
	assert := assert.New(t)
	var ts *time.Time
	ts = parseTime("2020-03-30T08:35:13+00:00")
	assert.NotNil(ts)
	assert.Equal(ts.Format(time.RFC3339), "2020-03-30T08:35:13Z")

}

func readMeta(html string) (*pipeline.Task, error) {
	t := &pipeline.Task{}
	t.SetHTML(html)

	return t, ReadMetadata(nil, t)
}

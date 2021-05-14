package metadata

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/akeil/scrapen/internal/pipeline"
)

func TestCountWords(t *testing.T) {
	html := `<html><head></head><body>
    foo
    </body></html>`
	countWords(t, html, 1)

	html = `<html><head></head><body>
    <h1>Headline</h1>
        <p>Foo Bar</p>
        <img src="image.jpg" />
        <hr />
        <div>Foo Bar</div>
    </body></html>`
	countWords(t, html, 5)

	html = `<html><head></head><body>
        <!-- Empty -->
    </body></html>`
	countWords(t, html, 0)

	// Punctuation, stopwords and quotes
	html = `<html><head></head><body>
    <h1>Headline Nr. 5</h1>
        <p>Foo a Bar.</p>
        <div>Foo "Bar" Baz!</div>
    </body></html>`
	countWords(t, html, 9)
}

func countWords(t *testing.T, html string, expected int) {
	assert := assert.New(t)
	task := &pipeline.Task{}
	task.SetHTML(html)

	err := CountWords(context.TODO(), task)
	assert.Nil(err)
	assert.Equal(expected, task.WordCount)
}

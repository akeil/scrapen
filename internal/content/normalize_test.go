package content

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeSpace(t *testing.T) {
	assert := assert.New(t)

	d := doc("<p>unchanged</p>")
	normalizeSpace(d)
	assert.Equal("<p>unchanged</p>", str(d))

	d = doc("<p> with space </p>")
	normalizeSpace(d)
	assert.Equal("<p>with space</p>", str(d))

	d = doc("<p> with <em>some</em> space </p>")
	normalizeSpace(d)
	assert.Equal("<p>with <em>some</em> space</p>", str(d))
}

func TestDeduplicateTitle(t *testing.T) {
	assert := assert.New(t)

	d := doc("<h2>My Title</h2><p>unchanged</p>")
	deduplicateTitle(d, "My Title")
	assert.Equal("<p>unchanged</p>", str(d))

	// case insensitive
	d = doc("<h2>MY TITLE</h2><p>unchanged</p>")
	deduplicateTitle(d, "My Title")
	assert.Equal("<p>unchanged</p>", str(d))

	// keep other headings
	d = doc("<h2>My Title</h2><p>unchanged</p> <h3>Other Title</h3>")
	deduplicateTitle(d, "My Title")
	assert.Equal("<p>unchanged</p> <h3>Other Title</h3>", str(d))
}

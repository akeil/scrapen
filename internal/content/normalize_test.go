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

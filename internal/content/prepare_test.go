package content

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnwrapNoscript(t *testing.T) {
	assert := assert.New(t)

	d := doc("<p>foo <noscript><em>bar</em></noscript> baz</p>")
	unwrapNoscript(d)
	assert.Equal("<p>foo <em>bar</em> baz</p>", str(d))
}

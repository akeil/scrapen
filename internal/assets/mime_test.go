package assets

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileExt(t *testing.T) {
	assert := assert.New(t)

	// the most important image types
	assert.Equal(".png", fileExt("image/png"))
	assert.Equal(".jpg", fileExt("image/jpeg"))
	assert.Equal(".gif", fileExt("image/gif"))
	assert.Equal(".svg", fileExt("image/svg+xml"))

	assert.Equal(".jpg", fileExt("image/jpg")) // non standard support
	assert.Equal(".png", fileExt("IMAGE/PNG")) // case insensitive

	assert.Equal("", fileExt("foo/bar")) // fallback
}

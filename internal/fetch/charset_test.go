package fetch

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/akeil/scrapen/internal/pipeline"
)

func TestCharsetByName(t *testing.T) {
	assert := assert.New(t)
	var h = http.Header{}
	var cs string

	task := &pipeline.Task{}

	cs = charsetFromHeader(task, h)
	assert.Equal("", cs)

	h.Add("Content-Type", "text/html")
	cs = charsetFromHeader(task, h)
	assert.Equal("", cs)

	h.Add("Content-Type", "text/plain")
	cs = charsetFromHeader(task, h)
	assert.Equal("", cs)

	h.Del("Content-Type")

	h.Add("Content-Type", "text/html; charset=iso-8859-1")
	cs = charsetFromHeader(task, h)
	assert.Equal("iso-8859-1", cs)

	h.Add("Content-Type", "text/html; charset=utf-8")
	cs = charsetFromHeader(task, h)
	assert.Equal("iso-8859-1", cs)
}

func TestDecoderByName(t *testing.T) {
	assert := assert.New(t)

	assert.NotNil(decoderByName("ISO 8859-1"))
	assert.NotNil(decoderByName("ISO-8859-1"))
	assert.NotNil(decoderByName("iso-8859-1"))
	assert.NotNil(decoderByName("iso 8859-1"))
}

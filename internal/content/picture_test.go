package content

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPicture(t *testing.T) {
	assert := assert.New(t)

	html := "<picture><img src=\"default.jpg\"/></picture>"
	d := doc(html)
	resolvePicture(d)
	assert.Equal(html, str(d))

	html = `<picture>
        <source srcset="large.png 200w, small.png 100w" />
        <source srcset="dense.png 2x, standard.png 1x" />
        <source srcset="" />
        <img src="default.jpg" />
    </picture>`
	d = doc(html)
	resolvePicture(d)
	assert.Equal("<img src=\"large.png\"/>", strings.TrimSpace(str(d)))
}

func TestAmpImg(t *testing.T) {
	assert := assert.New(t)

	html := `<figure>
		<amp-img src="image.jpg" srcset="one.jpg 704w, two.jpg 512w, three.jpg 310w">
			<img src="picture.jpg" srcset="large.jpg 704w, medium.jpg 512w, small.jpg 310w" />
		</amp-img>
		<figcaption>Caption</figcaption>
	</figure>`
	expect := `<figure>
		<img src="image.jpg" srcset="one.jpg 704w, two.jpg 512w, three.jpg 310w"/>
		<figcaption>Caption</figcaption>
	</figure>`
	d := doc(html)
	convertAmpImg(d)

	assert.Equal(expect, str(d))
}

func TestDataSrcset(t *testing.T) {
	assert := assert.New(t)

	html := "<picture><img src=\"default.jpg\"/></picture>"
	d := doc(html)
	resolvePicture(d)
	assert.Equal(html, str(d))

	html = `<picture>
        <source data-srcset="large.png 200w, small.png 100w" />
        <source data-srcset="dense.png 2x, standard.png 1x" />
        <source srcset="" />
        <img src="default.jpg" />
    </picture>`
	d = doc(html)
	resolvePicture(d)
	assert.Equal("<img src=\"large.png\"/>", strings.TrimSpace(str(d)))
}

func TestResolveSrcset(t *testing.T) {
	assert := assert.New(t)

	html := "<picture><img src=\"default.jpg\"/></picture>"
	d := doc(html)
	resolveSrcset(d)
	assert.Equal(html, str(d))

	d = doc(`<img src="foo.jpg" srcset="small.jpg 100w, large.jpg 200w"/>`)
	resolveSrcset(d)
	img := d.Selection.Find("img").First()
	src, _ := img.Attr("src")
	assert.Equal("large.jpg", src)
}

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

func TestUnwrapDivs(t *testing.T) {
	assert := assert.New(t)

	html := `
	<div>
      <div>
        <div>
          <picture class="ts-picture js-picture ts-picture--copytext-l">
            <source type="image/png" media="(max-width: 420px)" src="https://www.tagesschau.de/res/assets/image/lazy-image-placeholder.jpg" data-srcset="https://www.tagesschau.de/multimedia/bilder/alex-barck-101~_v-mittelgross1x1.jpg"/>
            <source type="image/png" media="(max-width: 767px)" src="https://www.tagesschau.de/res/assets/image/lazy-image-placeholder.jpg" data-srcset="https://www.tagesschau.de/multimedia/bilder/alex-barck-101~_v-videoweb1x1l.jpg"/>
            <img class="ts-image js-image" title="DJ Alex Barck legt am Dienstag (18.05.2010) beim Verve Club auf. | picture-alliance / Xamax" alt="DJ Alex Barck legt am Dienstag (18.05.2010) beim Verve Club auf. | picture-alliance / Xamax" src="https://www.tagesschau.de/multimedia/bilder/alex-barck-101~_v-videowebl.jpg"/>
          </picture>
        </div>
      </div>
      <div>
        <p>Sonst am Plattenteller, jetzt im Team des Impfzentrums Tegel: DJ Alex Barck.</p>
      </div>
    </div>`

	d := doc(html)
	unwrapDivs(d)
	assert.Equal(2, d.Find("div").Length())

}

func TestDropTemplates(t *testing.T) {
	assert := assert.New(t)

	d := doc(`
	<p>One</p>
	<template>some content</template>
	<p>Two</p>
	<template>
		<p>Template with HTML</p>
	</template>
	<p>Three</p>
	`)
	doPrepare(d)
	assert.Equal(0, d.Find("template").Length())
	assert.Equal(3, d.Find("p").Length())

	d = doc(`
	<p>One</p>
	<amp-list>
		<template>
			<p>{{placeholder}}</p>
		</template>
	</amp-list>
	<p>Two</p>
	<amp-ad>something</amp-ad>
	<p>Three</p>
	`)
	doPrepare(d)
	assert.Equal(3, d.Find("p").Length())
	assert.Equal(0, d.Find("amp-list").Length())
	assert.Equal(0, d.Find("amp-ad").Length())
	assert.Equal(0, d.Find("amp-template").Length())
}

func TestDropNavList(t *testing.T) {
	assert := assert.New(t)

	// this list should be dropped
	d := doc(`<p>head</p><ul><li><a>link</a></li><li><a>link 2</a></li></ul><p>tail</p>`)
	dropNavLists(d)
	assert.Equal(`<p>head</p><p>tail</p>`, str(d))

	// this list should be kept
	d = doc(`<p>head</p><ul><li><a>link</a></li><li>Not a link</li><li>Also not a link</li></ul><p>tail</p>`)
	dropNavLists(d)
	assert.Equal(`<p>head</p><ul><li><a>link</a></li><li>Not a link</li><li>Also not a link</li></ul><p>tail</p>`, str(d))
}

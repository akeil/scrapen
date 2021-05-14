package pipeline

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Pipeline func(ctx context.Context, t *Task) error

type Store interface {
	Put(k, contentType string, data []byte) error
	Get(k string) (string, []byte, error)
}

// Note: Store is the same interface as defined in the main scrapen module.
// The definition here is required for internal use.

type Task struct {
	ID           string
	URL          string
	ActualURL    string
	CanonicalURL string
	StatusCode   int
	Title        string
	Retrieved    time.Time
	Description  string
	PubDate      *time.Time
	Site         string
	SiteScheme   string
	Author       string
	ImageURL     string
	Images       []ImageInfo
	Feeds        []FeedInfo
	Enclosures   []Enclosure
	WordCount    int
	Store        Store
	document     *goquery.Document
	altDocument  *goquery.Document
	AltURL       string
	mx           sync.Mutex
}

func NewTask(s Store, id, url string) *Task {
	return &Task{
		ID:        id,
		URL:       url,
		Retrieved: time.Now().UTC(),
		Store:     s,
		Images:    make([]ImageInfo, 0),
	}
}

// Document returns the HTML content of this task as a DOM document.
// The document can be edited in place, i.e. all changes made to the document
// directly affect the task content.
func (t *Task) Document() *goquery.Document {
	return t.document
}

// HTML returns the HTML content for this task as a string.
func (t *Task) HTML() string {
	if t.document == nil {
		return ""
	}

	html, err := t.document.Selection.Find("html").First().Html()
	if err != nil {
		// TODO: log? panic?
		return ""
	}
	return html
}

// SetHTML sets the given string as the new HTML content.
// This will invalidate previous references to the goquery document.
// If you need a DOM document, call `Document()` again to retrieve one.
func (t *Task) SetHTML(s string) {
	r := strings.NewReader(s)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		// TODO: log? panic?
		return
	}
	t.document = doc
}

// SetAltHTML sets the alternate HTML content.
func (t *Task) SetAltHTML(s string) {
	r := strings.NewReader(s)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		// TODO: log? panic?
		return
	}
	t.altDocument = doc
}

// AltDocument is the altrnate HTML document.
func (t *Task) AltDocument() *goquery.Document {
	return t.altDocument
}

// TODO: still needed?
func (t *Task) PutAsset(k, contentType string, data []byte) error {
	return t.Store.Put(k, contentType, data)
}

func (t *Task) GetAsset(k string) (string, []byte, error) {
	return t.Store.Get(k)
}

func (t *Task) AddImage(i ImageInfo, data []byte) error {
	err := t.Store.Put(i.Key, i.ContentType, data)
	if err != nil {
		return err
	}

	t.mx.Lock()
	defer t.mx.Unlock()

	t.Images = append(t.Images, i)
	return nil
}

func (t *Task) AddEnclosure(e Enclosure) {
	t.mx.Lock()
	defer t.mx.Unlock()
	if t.Enclosures == nil {
		t.Enclosures = make([]Enclosure, 0)
	}
	t.Enclosures = append(t.Enclosures, e)
}

// ContentURL is the "best" URL for an item.
//
// If available, the actual URL is returned. Otherwise, the requested URL is used.
func (t *Task) ContentURL() string {
	if t.ActualURL != "" {
		return t.ActualURL
	} else if t.CanonicalURL != "" {
		return t.CanonicalURL
	}
	return t.URL
}

// ResolveURL resolves the given URL(-fragment) against the content URL that
// was determined for this task.
func (t *Task) ResolveURL(href string) (string, error) {
	b, err := url.Parse(t.ContentURL())
	if err != nil {
		return "", err
	}

	h, err := url.Parse(href)
	if err != nil {
		return "", err
	}

	return b.ResolveReference(h).String(), nil
}

type ImageInfo struct {
	Key         string
	ContentURL  string
	OriginalURL string
	ContentType string
}

func BuildPipeline(f ...Pipeline) Pipeline {
	return func(ctx context.Context, t *Task) error {
		var err error
		for _, p := range f {
			err = p(ctx, t)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

type memoryStore struct {
	assets map[string]asset
	mx     sync.Mutex
}

func NewMemoryStore() Store {
	return &memoryStore{
		assets: make(map[string]asset),
	}
}

func (m *memoryStore) Put(k, contentType string, data []byte) error {
	m.mx.Lock()
	defer m.mx.Unlock()

	m.assets[k] = asset{contentType: contentType, data: data}
	return nil
}

func (m *memoryStore) Get(k string) (string, []byte, error) {
	m.mx.Lock()
	defer m.mx.Unlock()

	asset, ok := m.assets[k]
	if !ok {
		return "", nil, fmt.Errorf("no asset with id %q", k)
	}
	return asset.contentType, asset.data, nil
}

type asset struct {
	contentType string
	data        []byte
}

const storePrefix = "store://"

// StoreURL builds a "store://" URL for the given store ID.
func StoreURL(id string) string {
	return storePrefix + id
}

// ParseStoreID extracts the store ID from a "store://" URL.
// Returns an empty string if this is not a store URL.
func ParseStoreID(url string) string {
	if strings.HasPrefix(url, storePrefix) {
		return strings.TrimPrefix(url, storePrefix)
	}
	return ""
}

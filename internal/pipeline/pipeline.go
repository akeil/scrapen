package pipeline

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"
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
	HTML         string
	Title        string
	Retrieved    time.Time
	Description  string
	PubDate      *time.Time
	Site         string
	SiteScheme   string
	Author       string
	ImageURL     string
	FeedInfo     []FeedInfo
	Store        Store
}

func NewTask(s Store, id, url string) *Task {
	return &Task{
		ID:        id,
		URL:       url,
		Retrieved: time.Now().UTC(),
		Store:     s,
	}
}

func (t *Task) PutAsset(k, contentType string, data []byte) error {
	return t.Store.Put(k, contentType, data)
}

func (t *Task) GetAsset(k string) (string, []byte, error) {
	return t.Store.Get(k)
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
}

func NewMemoryStore() Store {
	return &memoryStore{
		assets: make(map[string]asset),
	}
}

func (m *memoryStore) Put(k, contentType string, data []byte) error {
	m.assets[k] = asset{contentType: contentType, data: data}
	return nil
}

func (m *memoryStore) Get(k string) (string, []byte, error) {
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

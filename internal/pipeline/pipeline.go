package pipeline

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type Pipeline func(ctx context.Context, i *Item) (*Item, error)

type Store interface {
	Put(k, contentType string, data []byte) error
	Get(k string) (string, []byte, error)
}

type Item struct {
	URL          string
	ActualURL    string
	CanonicalURL string
	HTML         string
	Title        string
	Retrieved    time.Time
	Description  string
	PubDate      *time.Time
	Site         string
	Author       string
	ImageURL     string
	store        Store
}

func NewItem(s Store, url string) *Item {
	return &Item{
		URL:       url,
		Retrieved: time.Now().UTC(),
		store:     s,
	}
}

func (i *Item) PutAsset(k, contentType string, data []byte) error {
	return i.store.Put(k, contentType, data)
}

func (i *Item) GetAsset(k string) (string, []byte, error) {
	return i.store.Get(k)
}

// ContentURL is the "best" URL for an item.
//
// If available, the actual URL is returned. Otherwise, the requested URL is used.
func (i *Item) ContentURL() string {
	if i.ActualURL != "" {
		return i.ActualURL
	} else if i.CanonicalURL != "" {
		return i.CanonicalURL
	}
	return i.URL
}

func BuildPipeline(f ...Pipeline) Pipeline {
	return func(ctx context.Context, i *Item) (*Item, error) {
		var err error
		for _, p := range f {
			i, err = p(ctx, i)
			if err != nil {
				return i, err
			}
		}
		return i, nil
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

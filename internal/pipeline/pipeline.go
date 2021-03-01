package pipeline

import (
	"context"
	"fmt"
	"time"
)

type Pipeline func(ctx context.Context, i *Item) (*Item, error)

type Store interface {
	Put(k, contentType string, data []byte) error
	Get(k string) (string, []byte, error)
}

type Item struct {
	URL          string
	CanonicalURL string
	HTML         string
	Title        string
	Retrieved    time.Time
	Description  string
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

func (i *Item) Copy() *Item {
	return &Item{
		URL:       i.URL,
		HTML:      i.HTML,
		Title:     i.Title,
		Retrieved: i.Retrieved,
		store:     i.store,
	}
}

func (i *Item) PutAsset(k, contentType string, data []byte) error {
	return i.store.Put(k, contentType, data)
}

func (i *Item) GetAsset(k string) (string, []byte, error) {
	return i.store.Get(k)
}

func BuildPipeline(f ...Pipeline) Pipeline {
	return func(ctx context.Context, i *Item) (*Item, error) {
		item := i.Copy()
		var err error
		for _, p := range f {
			item, err = p(ctx, item)
			if err != nil {
				return item, err
			}
		}
		return item, nil
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

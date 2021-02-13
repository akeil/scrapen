package pipeline

import (
	"context"
	"fmt"
	"time"
)

type Pipeline func(ctx context.Context, i *Item) (*Item, error)

type Item struct {
	Url       string
	Html      string
	Title     string
	Retrieved time.Time
	assets    map[string]asset
}

type asset struct {
	contentType string
	data        []byte
}

func NewItem(url string) *Item {
	return &Item{
		Url:       url,
		Retrieved: time.Now().UTC(),
		assets:    make(map[string]asset),
	}
}

func (i *Item) Copy() *Item {
	return &Item{
		Url:       i.Url,
		Html:      i.Html,
		Title:     i.Title,
		Retrieved: i.Retrieved,
		assets:    i.assets,
	}
}

func (i *Item) PutAsset(id, contentType string, data []byte) error {
	i.assets[id] = asset{contentType: contentType, data: data}
	return nil
}

func (i *Item) GetAsset(id string) (string, []byte, error) {
	asset, ok := i.assets[id]
	if !ok {
		return "", nil, fmt.Errorf("no asset with id %q", id)
	}
	return asset.contentType, asset.data, nil
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

package handler

import (
	"context"
	"strings"

	"github.com/heyLu/lp/go/things/storage"
)

var _ Handler = SearchHandler{}

type SearchHandler struct{}

func (s SearchHandler) CanHandle(input string) (string, bool) {
	return "search", strings.HasPrefix(input, "search")
}

func (s SearchHandler) Parse(input string) (Thing, error) {
	return Search(input), nil
}

func (s SearchHandler) Query(ctx context.Context, db storage.Storage, namespace string, input string) (storage.Rows, error) {
	// TODO: implement actual search :)
	return db.Query(ctx, namespace)
}

func (s SearchHandler) Render(ctx context.Context, row *storage.Row) (Renderer, error) {
	if row.Kind == "search" || row.Kind == "overview" {
		return StringRenderer("searching..."), nil
	}

	_, handler := All.For(row.Kind)

	return handler.Render(ctx, row)
}

type Search string

func (h Search) ToRow() *storage.Row {
	return &storage.Row{
		Metadata: storage.Metadata{
			Kind: "search",
		},
		Summary: string(h),
	}
}

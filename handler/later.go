package handler

import (
	"context"
	"html/template"
	"strings"

	"github.com/heyLu/lp/go/things/storage"
)

var _ Handler = LaterHandler{}

type LaterHandler struct{}

func (nh LaterHandler) CanHandle(input string) (string, bool) {
	return "later", strings.HasPrefix(input, "later")
}

func (nh LaterHandler) Parse(input string) (Thing, error) {
	idx := strings.Index(input, " ")
	if idx == -1 {
		idx = len(input)
	}
	content := input[idx:]

	later := Later{
		Row: &storage.Row{
			Metadata: storage.Metadata{
				Kind: "later",
			},
			Summary: content,
		},
	}

	return later, nil
}

func (nh LaterHandler) Query(ctx context.Context, db storage.Storage, namespace string, input string) (storage.Rows, error) {
	parts := strings.SplitN(input, " ", 2)
	if len(parts) == 1 {
		return db.Query(ctx, namespace, storage.Kind("later"))
	}
	return db.Query(ctx, namespace, storage.Kind("later"), storage.Match("summary", parts[1]))
}

func (nh LaterHandler) Render(ctx context.Context, row *storage.Row) (Renderer, error) {
	return TemplateRenderer{
		Template: laterTemplate,
		Data:     Later{Row: row},
	}, nil
}

type Later struct {
	*storage.Row
}

func (n Later) ToRow() *storage.Row { return n.Row }

var laterTemplate = template.Must(template.Must(commonTemplates.Clone()).Parse(`
{{ define "content" }}
<div>{{ markdown .Summary }}</div>
{{ end }} 
`))

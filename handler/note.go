package handler

import (
	"context"
	"html/template"
	"regexp"
	"strings"

	"github.com/heyLu/lp/go/things/storage"
)

var _ Handler = NoteHandler{}

type NoteHandler struct{}

func (nh NoteHandler) CanHandle(input string) (string, bool) {
	return "note", strings.HasPrefix(input, "note")
}

var urlRe = regexp.MustCompile(`(\w+)://[^ ]+`)

func (nh NoteHandler) Parse(input string) (Thing, error) {
	idx := strings.Index(input, " ")
	if idx == -1 {
		idx = len(input)
	}
	content := input[idx:]

	note := Note{
		Row: &storage.Row{
			Metadata: storage.Metadata{
				Kind: "note",
			},
			Summary: content,
		},
	}

	about := urlRe.FindString(content)
	if about != "" {
		note.Ref.String = about
	}

	return note, nil
}

func (nh NoteHandler) Query(ctx context.Context, db storage.Storage, namespace string, input string) (storage.Rows, error) {
	parts := strings.SplitN(input, " ", 2)
	if len(parts) == 1 {
		return db.Query(ctx, namespace, storage.Kind("note"))
	}
	return db.Query(ctx, namespace, storage.Kind("note"), storage.Match("summary", parts[1]))
}

func (nh NoteHandler) Render(ctx context.Context, row *storage.Row) (Renderer, error) {
	return TemplateRenderer{
		Template: noteTemplate,
		Data:     Note{Row: row},
	}, nil
}

type Note struct {
	*storage.Row
}

func (n Note) ToRow() *storage.Row { return n.Row }

var noteTemplate = template.Must(template.Must(commonTemplates.Clone()).Parse(`
{{ define "content" }}
<div>{{ markdown .Summary }}</div>
{{ end }} 
`))

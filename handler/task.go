package handler

import (
	"context"
	"database/sql"
	"html/template"
	"strings"

	"github.com/heyLu/lp/go/things/storage"
)

var _ Handler = TaskHandler{}

type TaskHandler struct{}

func (nh TaskHandler) CanHandle(input string) (string, bool) {
	return "task", strings.HasPrefix(input, "task")
}

func (nh TaskHandler) Parse(input string) (Thing, error) {
	idx := strings.Index(input, " ")
	if idx == -1 {
		idx = len(input)
	}
	content := input[idx:]

	task := Task{
		Row: &storage.Row{
			Metadata: storage.Metadata{
				Kind: "task",
			},
			Summary: content,
			Bool: sql.NullBool{
				Bool:  false,
				Valid: true,
			},
		},
	}

	return task, nil
}

func (nh TaskHandler) Query(ctx context.Context, db storage.Storage, namespace string, input string) (storage.Rows, error) {
	parts := strings.SplitN(input, " ", 2)
	if len(parts) == 1 {
		return db.Query(ctx, namespace, storage.Kind("task"))
	}
	return db.Query(ctx, namespace, storage.Kind("task"), storage.Match("summary", parts[1]))
}

func (nh TaskHandler) Render(ctx context.Context, row *storage.Row) (Renderer, error) {
	return TemplateRenderer{
		Template: taskTemplate,
		Data:     Task{Row: row},
	}, nil
}

type Task struct {
	*storage.Row
}

func (n Task) ToRow() *storage.Row { return n.Row }

var taskTemplate = template.Must(template.Must(commonTemplates.Clone()).Parse(`
{{ define "content" }}
<div>{{ markdown .Summary }}</div>
<input type="checkbox" disabled{{ if .Bool.Bool }} checked{{ end }} />
{{ end }} 
`))

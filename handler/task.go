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
			// TODO: use Num as priority?
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
<div{{ if .Bool.Bool }} class="done"{{ end }}>
	<header>
		{{ if .Bool.Bool }}<s>{{ end }}
		<input type="checkbox"
			{{ if .Bool.Bool }} checked{{ end }}
			{{ if (gt .ID 0) }} hx-post="/{{ .Namespace }}/{{ .Kind }}/{{ .ID }}" hx-include="next [name='bool']"{{ end }}
			/>

		<h1>{{ markdown .Summary }}</h1>
		{{ if .Bool.Bool }}</s>{{ end }}
	</header>

	{{ markdown .Content.String }}

</div>
<input type="hidden" name="bool" value="{{ .Bool.Bool }}" />
{{ end }} 
`))

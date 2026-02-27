package handler

import (
	"context"
	"html/template"

	"github.com/heyLu/lp/go/things/storage"
)

var _ Handler = &GenericHandler{}

type GenericHandler struct {
}

// CanHandle implements [Handler].
func (g *GenericHandler) CanHandle(input string) (string, bool) {
	panic("unimplemented")
}

// Parse implements [Handler].
func (g *GenericHandler) Parse(input string) (Thing, error) {
	panic("unimplemented")
}

// Query implements [Handler].
func (g *GenericHandler) Query(ctx context.Context, db storage.Storage, namespace string, input string) (storage.Rows, error) {
	panic("unimplemented")
}

// Render implements [Handler].
func (g *GenericHandler) Render(ctx context.Context, row *storage.Row) (Renderer, error) {
	return TemplateRenderer{
		Template: genericTemplate,
		Data: Generic{
			Row: row,
		},
	}, nil
}

type Generic struct {
	*storage.Row
}

var genericTemplate = template.Must(template.Must(commonTemplates.Clone()).Parse(`
{{ define "content" }}
<form method="POST" action="">
  	<div>
  		<input name="summary" type="text" value="{{ .Summary }}" />
  	</div>
  	<div>
		<textarea name="content">{{ .Content.String }}</textarea>
	</div>
	<div>
		<input type="text" value="{{ .Tags }}" />
	</div>

	{{ if .Number.Valid }}
	<div>
		<input name="number" type="number" value="{{ .Number.Int64 }}" />
	</div>
	{{ end }}
	{{ if .Float.Valid }}
	<div>
		<input name="float" type="number" value="{{ .Number.Float }}" />
	</div>
	{{ end }}
	<div>
		<input name="bool-valid" type="hidden" value="{{ .Bool.Valid }}" />
		<input name="bool" type="checkbox" {{ if not .Bool.Valid }}disabled{{ end }} {{ if .Bool.Bool }}checked{{ end }} />
	</div>
	{{ if .Time.Valid }}
	<div>
		<input type="time" value="{{ .Time.Time }}" />
	</div>
	{{ end }}
	{{ if .Number.Valid }}
	<div>
		<input type="number" value="{{ .Number.Int64 }}" />
	</div>
	{{ end }}

	{{ range $k, $v := .Fields }}
	<div>
		<label for="{{ $k }}" />
		<input name="{{ $k }}" type="text" value="{{ $v }}" />
	</div>
	{{ end }}

	<div>
		<input type="submit" value="save" />
	</div>
</form>
{{ end }} 
`))

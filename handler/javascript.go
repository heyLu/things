package handler

import (
	"context"
	"html/template"
	"strings"

	"github.com/heyLu/lp/go/things/storage"
)

var _ Handler = JavaScriptHandler{}

type JavaScriptHandler struct{}

func (j JavaScriptHandler) CanHandle(input string) (string, bool) {
	return "javascript", strings.HasPrefix(input, "javascript") || strings.HasPrefix(input, "js")
}

func (j JavaScriptHandler) Parse(input string) (Thing, error) {
	parts := strings.SplitN(input, " ", 2)
	if len(parts) < 2 {
		return JavaScript("/* your code here ✨ */"), nil
	}
	return JavaScript(parts[1]), nil
}

func (j JavaScriptHandler) Query(ctx context.Context, db storage.Storage, namespace string, input string) (storage.Rows, error) {
	return db.Query(ctx, namespace, storage.Kind("javascript"))
}

func (j JavaScriptHandler) Render(ctx context.Context, row *storage.Row) (Renderer, error) {
	return TemplateRenderer{Template: javaScriptTemplate, Data: row.Summary}, nil
}

type JavaScript string

func (j JavaScript) ToRow() *storage.Row {
	return &storage.Row{
		Metadata: storage.Metadata{
			Kind: "javascript",
		},
		Summary: string(j),
	}
}

var javaScriptTemplate = template.Must(template.New("").Parse(`
<section class="thing js">
	<pre>function(canvas, ctx) {
<textarea name="summary" class="js-code">{{ . }}</textarea>
}</pre>

	<code><pre class="js-output"></pre></code>

	<canvas class="js-canvas"></canvas>
</section>
`))

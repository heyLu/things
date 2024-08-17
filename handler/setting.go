package handler

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"strings"

	"github.com/heyLu/lp/go/things/storage"
)

var _ Handler = HelpHandler{}

type SettingHandler struct{}

func (s SettingHandler) CanHandle(input string) (string, bool) {
	return "setting", strings.HasPrefix(input, "setting")
}

func (s SettingHandler) Parse(input string) (Thing, error) {
	parts := strings.SplitN(input, " ", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("usage: setting <key> <value...>")
	}

	return &Setting{
		Row: &storage.Row{
			Metadata: storage.Metadata{
				Kind: "setting",
			},
			Summary: parts[1],
			Content: sql.NullString{
				String: parts[2],
				Valid:  true,
			},
		},
	}, nil
}

func (s SettingHandler) Query(ctx context.Context, db storage.Storage, namespace string, input string) (storage.Rows, error) {
	return db.Query(ctx, namespace, storage.Kind("setting"))
}

func (s SettingHandler) Render(ctx context.Context, row *storage.Row) (Renderer, error) {
	return TemplateRenderer{Template: settingTemplate, Data: &Setting{Row: row}}, nil
}

type Setting struct {
	*storage.Row
}

func (s *Setting) ToRow() *storage.Row {
	return s.Row
}

var settingTemplate = template.Must(template.Must(commonTemplates.Clone()).Parse(`
{{ define "content" }}
{{ .Summary }}: {{ if (eq .Summary "namespace.token") }}********{{ else }}{{ .Content.String }}{{ end }}
{{ end }}
`))

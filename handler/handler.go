package handler

import (
	"bytes"
	"context"
	"fmt"
	"html"
	"html/template"
	"net/http"

	"github.com/heyLu/lp/go/things/storage"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

var All = Handlers([]Handler{
	// TODO: bookmark <url> note
	ReminderHandler{},
	TrackHandler{},
	NoteHandler{},
	LaterHandler{},
	TaskHandler{},
	ByDateHandler{},
	JavaScriptHandler{},
	SearchHandler{},
	SettingHandler{},
	MathHandler{},
	HelpHandler{},
	OverviewHandler{},
})

type Handlers []Handler

func (hs Handlers) For(kind string) (string, Handler) {
	for _, h := range hs {
		k, _ := h.CanHandle("")
		if k == kind {
			return kind, h
		}
	}

	// try if someone can handle it, e.g. if kind is 2024-08
	for _, h := range hs {
		_, ok := h.CanHandle(kind)
		if ok {
			return kind, h
		}
	}

	return "", nil
}

type Handler interface {
	CanHandle(input string) (string, bool)
	Parse(input string) (Thing, error)

	Query(ctx context.Context, db storage.Storage, namespace string, input string) (storage.Rows, error)
	Render(ctx context.Context, row *storage.Row) (Renderer, error)
}

type Thing interface {
	ToRow() *storage.Row
}

type Renderer interface {
	Render(ctx context.Context, w http.ResponseWriter) error
}

type StringRenderer string

func (sr StringRenderer) Render(ctx context.Context, w http.ResponseWriter) error {
	fmt.Fprintln(w, "<pre>"+html.EscapeString(string(sr))+"</pre")
	return nil
}

type HTMLRenderer string

func (hr HTMLRenderer) Render(ctx context.Context, w http.ResponseWriter) error {
	fmt.Fprintln(w, hr)
	return nil
}

type ListRenderer []Renderer

func (lr ListRenderer) Render(ctx context.Context, w http.ResponseWriter) error {
	fmt.Fprintln(w, "<ul>")
	for _, r := range lr {
		fmt.Fprintln(w, "<li>")
		err := r.Render(ctx, w)
		if err != nil {
			return err
		}
		fmt.Fprintln(w, "</li>")
	}
	fmt.Fprintln(w, "</ul>")
	return nil
}

type SequenceRenderer []Renderer

func (sr SequenceRenderer) Render(ctx context.Context, w http.ResponseWriter) error {
	for _, r := range sr {
		err := r.Render(ctx, w)
		if err != nil {
			return err
		}
	}
	return nil
}

type TemplateRenderer struct {
	*template.Template

	Metadata *storage.Metadata
	Data     any
}

func (tr TemplateRenderer) Render(ctx context.Context, w http.ResponseWriter) error {
	return tr.Template.ExecuteTemplate(w, "thing", tr.Data)
}

var commonMarkdown = goldmark.New(goldmark.WithExtensions(extension.GFM))

var commonFuncs = template.FuncMap{
	// TODO: linkify tags
	"markdown": func(md string) (template.HTML, error) {
		buf := new(bytes.Buffer)
		err := commonMarkdown.Convert([]byte(md), buf)
		if err != nil {
			return "", err
		}
		return template.HTML(buf.String()), nil
	},
}

var commonTemplates = template.Must(template.New("").Funcs(commonFuncs).Parse(`
{{ define "thing" }}
<section class="thing {{ .Kind }}">
	<div class="content">
	{{ template "content" . }}
	</div>

	<footer class="meta">
		<time class="date-created" time="{{ .DateCreated }}" title="{{ .DateCreated }}">{{ .DateCreated.Format "2006-01-02 15:04:05" }}</time>

		<span class="tags">{{ range .Tags }}{{ if (gt (len .) 1) }}<a href="/tag/{{ slice . 1 }}">{{ . }}</a> {{ end }}{{ end }}
	</footer>
</section>
{{ end }}
`))

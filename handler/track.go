package handler

import (
	"context"
	"fmt"
	"html/template"
	"math"
	"strconv"
	"strings"

	"github.com/heyLu/lp/go/things/storage"
)

var _ Handler = TrackHandler{}

type TrackHandler struct{}

func (th TrackHandler) CanHandle(input string) (string, bool) {
	return "track", strings.HasPrefix(input, "track")
}

func (th TrackHandler) Parse(input string) (Thing, error) {
	var t Track
	t.Row = &storage.Row{Metadata: storage.Metadata{Kind: "track"}}

	parts := strings.SplitN(input, " ", 4)
	if len(parts) > 1 {
		t.Summary = parts[1]
	}

	if len(parts) > 2 {
		num, err := strconv.ParseFloat(parts[2], 64)
		if err != nil {
			return nil, err
		}
		t.Float.Float64 = num
		t.Float.Valid = true

	}

	if len(parts) > 3 {
		t.Content.String = parts[3]
		t.Content.Valid = true
	}

	return &t, nil
}

func (th TrackHandler) Query(ctx context.Context, db storage.Storage, namespace string, input string) (storage.Rows, error) {
	thing, err := th.Parse(input)
	if err != nil {
		return nil, err
	}

	track := thing.(*Track)

	if track.Summary == "" {
		return db.Query(ctx, namespace, storage.Kind(track.Kind))
	}

	return db.Query(ctx, namespace, storage.Kind(track.Kind), storage.Summary(track.Summary))
}

func (th TrackHandler) Render(ctx context.Context, row *storage.Row) (Renderer, error) {
	return TemplateRenderer{Template: trackTemplate, Data: &Track{Row: row}}, nil
}

var _ Thing = &Track{}

type Track struct{ *storage.Row }

func (t *Track) Category() string { return t.Summary }
func (t *Track) Num() *float64 {
	if t.Float.Valid {
		return &t.Float.Float64
	} else {
		return nil
	}
}
func (t *Track) Notes() string { return t.Content.String }

func (t *Track) FormatValue() string {
	i, f := math.Modf(t.Float.Float64)
	switch t.Summary {
	case "sport":
		if f < 0.01 {
			return fmt.Sprintf("%dmin", int(i))
		}
		return fmt.Sprintf("%d:%02dmin", int(i), int(f*100))
	case "sleep":
		return fmt.Sprintf("%.2fhrs", t.Float.Float64)
	case "ready", "up", "bed":
		return fmt.Sprintf("%d:%02dhrs", int(i), int(f*100))
	case "groceries":
		return fmt.Sprintf("%.2feur", t.Float.Float64)
	case "weight":
		return fmt.Sprintf("%.2fkg", t.Float.Float64)
	default:
		if f < 0.001 {
			return fmt.Sprintf("%d", int(i))
		}
		return fmt.Sprintf("%f", t.Float.Float64)
	}
}

func (t *Track) ToRow() *storage.Row { return t.Row }

var trackTemplate = template.Must(template.Must(commonTemplates.Clone()).Parse(`
{{ define "content" }}
{{ .Category }}
{{ if .Num }}
<span{{ if (eq .Category "mood") }} style="opacity: calc({{ .Num }}/100)"{{ end }}>
	{{ .FormatValue }}
</span>
{{ end }}
{{ if .Notes }}<p>{{ .Notes | markdown }}</p>{{ end }}
{{ end }}
`))

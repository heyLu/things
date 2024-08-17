package main

import (
	"context"
	"crypto/rand"
	"embed"
	"errors"
	"flag"
	"fmt"
	"html"
	"log"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/heyLu/lp/go/things/handler"
	"github.com/heyLu/lp/go/things/storage"
)

var settings struct {
	Addr   string
	DBPath string
}

//go:embed static
var staticFS embed.FS

func main() {
	flag.StringVar(&settings.Addr, "addr", "localhost:5000", "Address to listen on")
	flag.StringVar(&settings.DBPath, "db-path", "things.db", "Path to db file")
	flag.Parse()

	dbStorage, err := storage.NewDBStorage(context.Background(), "file:"+settings.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer dbStorage.Close()

	things := &Things{
		handlers: handler.All,
		storage:  dbStorage,
	}

	things.kinds = make(map[string]bool, len(things.handlers))
	for _, h := range things.handlers {
		kind, _ := h.CanHandle("")
		if kind == "" {
			log.Fatalf("invalid handler %#v", h)
		}
		things.kinds[kind] = true
	}

	router := chi.NewRouter()

	namespaceMiddleware := NamespaceMiddleware{router: router, kinds: things.kinds}
	tokenMiddleware := tokenMiddleware{dbStorage}
	router.Use(
		namespaceMiddleware.Middleware,
		tokenMiddleware.Middleware,
	)

	router.Get("/*", things.HandleIndex)

	router.Get("/token", tokenMiddleware.HandleToken)
	router.Post("/token", tokenMiddleware.SetToken)

	router.Route("/{namespace}", func(namespaceRouter chi.Router) {
		namespaceRouter.Use(namespaceMiddleware.Middleware)

		namespaceRouter.Get("/thing", things.HandleThing)
		namespaceRouter.Post("/thing", things.HandleThing)

		namespaceRouter.Get("/{kind}", things.HandleList)
		namespaceRouter.Get("/{kind}/{category}", things.HandleList)
		namespaceRouter.Get("/{kind}/{category}/{id}", things.HandleFind)
	})

	router.Get("/{kind}", func(w http.ResponseWriter, req *http.Request) {
		// check if {kind} param is a valid kind, render a namespace index if not, e.g. to serve /fun-stuff as fun-stuff namespace
		kind := chi.URLParam(req, "kind")
		if _, ok := things.kinds[kind]; kind != "" && !ok {
			things.HandleIndex(w, req.WithContext(context.WithValue(req.Context(), NamespaceKey, kind)))
			return
		}

		things.HandleList(w, req)
	})

	router.Handle("/static/*", http.FileServerFS(staticFS))

	log.Printf("Listening on http://%s", settings.Addr)
	log.Fatal(http.ListenAndServe(settings.Addr, router))
}

type Things struct {
	handlers handler.Handlers
	kinds    map[string]bool

	storage storage.Storage
}

type Handler func(ctx context.Context, storage storage.Storage, namespace string, w http.ResponseWriter, input string, save bool) error

var ErrNotHandled = errors.New("not handled")

func (t *Things) HandleIndex(w http.ResponseWriter, req *http.Request) {
	pageWithContent(w, req, "", nil)
}

func pageWithContent(w http.ResponseWriter, req *http.Request, input string, content handler.Renderer) {
	namespace := req.Context().Value(NamespaceKey).(string)

	fmt.Fprintf(w, `<!doctype html>
<html>
<head>
	<meta charset="utf-8" />
	<meta name="viewport" content="width=device-width,minimum-scale=1,initial-scale=1" />
	<title>things</title>

	<link rel="stylesheet" href="/static/things.css" />
	<link rel="icon" href="data:image/svg+xml,<svg xmlns=%%22http://www.w3.org/2000/svg%%22 viewBox=%%220 0 100 100%%22><text y=%%22.9em%%22 font-size=%%2290%%22>🐦‍⬛</text></svg>" />
</head>

<body>
	<main>
		<form hx-post="/%s/thing" hx-target="#answer" hx-indicator="#waiting">
			<input id="tell-me" name="tell-me" type="text" autofocus autocomplete="off" placeholder="tell me things"
				value=%q
				hx-get="/%s/thing"
				hx-trigger="input changed delay:250ms"
				hx-target="#answer"
				hx-indicator="#waiting" />
			<input name="save" value="yes" hidden />
			<input type="submit" value="💾" />
		    <img id="waiting" class="htmx-indicator" src="/static/three-dots.svg" />
	    </form>

		<section id="answer">`,
		url.PathEscape(namespace),
		input,
		url.PathEscape(namespace),
	)

	if content != nil {
		content.Render(req.Context(), w)
	}

	fmt.Fprintf(w, `
		</section>

	</main>

	<footer class="info">
		<span id="namespace">namespace: %s</span>
	</footer>

	<script src="/static/htmx.min.js"></script>
	<script src="/static/things.js"></script>
</body>
</html>`,
		html.EscapeString(namespace),
	)
}

func (t *Things) HandleThing(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		http.Error(w, "could not parse form", http.StatusBadRequest)
		return
	}

	tellMe := req.Form.Get("tell-me")

	ctx, cancel := context.WithTimeout(req.Context(), 1*time.Second)
	defer cancel()

	save := req.Method == http.MethodPost

	handled := false
	for _, handler := range t.handlers {
		err := t.handle(ctx, handler, t.storage, w, tellMe, save)
		if err == ErrNotHandled {
			continue
		}

		handled = true

		if err != nil {
			fmt.Fprintln(w, html.EscapeString(err.Error()))
		}

		break
	}

	if !handled {
		fmt.Fprintln(w, "don't know what to do with that (yet)")
	}
}

func (t *Things) handle(ctx context.Context, hndl handler.Handler, storage storage.Storage, w http.ResponseWriter, input string, save bool) error {
	kind, ok := hndl.CanHandle(input)
	if !ok {
		return ErrNotHandled
	}

	fmt.Fprintln(w, kind)

	thing, err := hndl.Parse(input)
	if err != nil {
		return err
	}

	row := thing.(handler.Thing).ToRow()
	row.Namespace = ctx.Value(NamespaceKey).(string)

	if save {
		err := storage.Insert(ctx, row)
		if err != nil {
			return err
		}

		fmt.Fprintln(w, "saved!")
	}

	seq := []handler.Renderer{}

	if row.Summary != "" {
		// TODO: thing.CanSave and only then preview?
		previewRenderer, err := hndl.(handler.Handler).Render(ctx, row)
		if err != nil {
			return err
		}

		seq = append(seq,
			previewRenderer,
			handler.HTMLRenderer("<hr />"),
		)
	}

	listRenderer, err := t.renderList(ctx, hndl.(handler.Handler), row.Namespace, input)
	if err != nil {
		return err
	}
	seq = append(seq, listRenderer)

	renderer := handler.SequenceRenderer(seq)
	return renderer.Render(ctx, w)
}

func (t *Things) HandleList(w http.ResponseWriter, req *http.Request) {
	kindParam := chi.URLParam(req, "kind")
	kind, hndl := t.handlers.For(kindParam)
	if hndl == nil {
		http.Error(w, "unknown kind "+kindParam, http.StatusNotFound)
		return
	}

	// args := n.QueryArgs(make([]any, 0, 1)) // TODO: filter by category/first param

	input := kind

	namespace := req.Context().Value(NamespaceKey).(string)

	renderer, err := t.renderList(req.Context(), hndl.(handler.Handler), namespace, input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pageWithContent(w, req, input, renderer)
}

func (t *Things) renderList(ctx context.Context, hndl handler.Handler, namespace string, input string) (handler.Renderer, error) {
	rows, err := hndl.Query(ctx, t.storage, namespace, input)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := []handler.Renderer{}
	for rows.Next() {
		var row storage.Row
		err := rows.Scan(&row)
		if err != nil {
			return nil, err
		}

		renderer, err := hndl.Render(ctx, &row)
		if err != nil {
			return nil, err
		}

		res = append(res, renderer)
	}

	return handler.ListRenderer(res), nil
}

func (t *Things) HandleFind(w http.ResponseWriter, req *http.Request) {
	http.Error(w, "not implemented", http.StatusInternalServerError)
}

var NamespaceKey struct{}
var NamespaceCookieName = "things_namespace"

type NamespaceMiddleware struct {
	router chi.Router
	kinds  map[string]bool
}

func (nm *NamespaceMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/token" || strings.HasPrefix(req.URL.Path, "/static/") {
			next.ServeHTTP(w, req)
			return
		}

		routeCtx := chi.NewRouteContext()
		nm.router.Match(routeCtx, req.Method, req.URL.Path)

		namespace := routeCtx.URLParam("namespace")
		if _, ok := nm.kinds[routeCtx.URLParam("kind")]; namespace == "" && !ok {
			namespace = routeCtx.URLParam("kind")
		}
		if namespace != "" {
			ctx := context.WithValue(req.Context(), NamespaceKey, namespace)
			next.ServeHTTP(w, req.WithContext(ctx))
			return
		}

		namespaceCookie, err := req.Cookie(NamespaceCookieName)
		if err != nil && err != http.ErrNoCookie {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err == http.ErrNoCookie {
			// http.Redirect(w, req, "/new-namespace", http.StatusSeeOther)
			// return

			// TODO: only set on first save? 🤔

			ns := make([]byte, 8)
			_, err := rand.Read(ns)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			namespaceCookie = &http.Cookie{
				Name:  NamespaceCookieName,
				Value: fmt.Sprintf("%x", ns),
			}
		}

		// set cookie again to refresh it
		namespaceCookie.Path = "/"
		namespaceCookie.MaxAge = 60 * 60 * 24 * 365
		namespaceCookie.SameSite = http.SameSiteStrictMode
		http.SetCookie(w, namespaceCookie)

		namespace = namespaceCookie.Value

		ctx := context.WithValue(req.Context(), NamespaceKey, namespace)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

type tokenMiddleware struct {
	storage.Storage
}

var TokenCookieName = "things_namespace_token"

func (s tokenMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/token" || strings.HasPrefix(req.URL.Path, "/static/") {
			next.ServeHTTP(w, req)
			return
		}

		namespace := req.Context().Value(NamespaceKey).(string)
		tokens, err := s.getTokens(req.Context(), namespace)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(tokens) == 0 {
			next.ServeHTTP(w, req)
			return
		}

		tokenCookie, err := req.Cookie(TokenCookieName)
		if err != nil && err != http.ErrNoCookie {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err == http.ErrNoCookie || !slices.Contains(tokens, tokenCookie.Value) {
			http.Redirect(w, req, "/token?redirect-to="+url.QueryEscape(req.URL.Path), http.StatusSeeOther)
			return
		}

		// set cookie again to refresh it
		tokenCookie.Path = "/"
		tokenCookie.MaxAge = 60 * 60 * 24 * 365
		tokenCookie.SameSite = http.SameSiteStrictMode
		http.SetCookie(w, tokenCookie)

		next.ServeHTTP(w, req)
	})
}

func (s tokenMiddleware) getTokens(ctx context.Context, namespace string) ([]string, error) {
	rows, err := s.Query(ctx, namespace, storage.Kind("setting"), storage.Summary("namespace.token"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tokens := make([]string, 0, 1)
	for rows.Next() {
		var row storage.Row
		err := rows.Scan(&row)
		if err != nil {
			return nil, err
		}

		if !row.Content.Valid {
			return nil, fmt.Errorf("no value for namespace.token")
		}
		tokens = append(tokens, row.Content.String)
	}

	return tokens, nil
}

func (s tokenMiddleware) HandleToken(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, `
<!doctype html>
<html>
<head>
	<meta charset="utf-8" />
	<meta name="viewport" content="width=device-width,minimum-scale=1,initial-scale=1" />
	<title>things</title>

	<link rel="stylesheet" href="/static/things.css" />
	<link rel="icon" href="data:image/svg+xml,<svg xmlns=%%22http://www.w3.org/2000/svg%%22 viewBox=%%220 0 100 100%%22><text y=%%22.9em%%22 font-size=%%2290%%22>🐦‍⬛</text></svg>" />
</head>

<body>
	<main>
		<form target="/token?redirect-to=%s" method="POST">
			<input name="token" type="text" size="50" placeholder="token" />
			<input type="submit" value="Set token" />
		</form>

		<hr />

		<form target="/token?redirect-to=%s" method="POST">
			<input type="hidden" name="delete" value="delete" />
			<input type="submit" value="Delete token" />
		</form>
	</main>
</body>
</html>`,
		html.EscapeString(req.URL.Query().Get("redirect-to")),
		html.EscapeString(req.URL.Query().Get("redirect-to")),
	)
}

func (s tokenMiddleware) SetToken(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token := req.FormValue("token")
	cookie := &http.Cookie{
		Name:     TokenCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 365,
		SameSite: http.SameSiteStrictMode,
	}

	if req.FormValue("delete") == "delete" {
		cookie.Value = ""
		cookie.MaxAge = -1
	}

	http.SetCookie(w, cookie)

	redirectTo := req.URL.Query().Get("redirect-to")
	if redirectTo == "" {
		redirectTo = "/"
	}
	http.Redirect(w, req, redirectTo, http.StatusSeeOther)
}

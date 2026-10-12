// Command i41eu-web serves i41.eu: one page about the i41 home platform, with links to its
// parts (s.i41.eu, m.i41.eu, the apps on q41.eu) and to the code. It also serves s.i41.eu, the
// Stackchan robot's page with its pictures (/s/*.png, /s/*.webp) and a page per app
// (s.i41.eu/<app>). The pages are embedded and chosen by the request's host.
package main

import (
	"bytes"
	"embed"
	"flag"
	"html/template"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

//go:embed index.html
var indexHTML []byte

//go:embed s.html
var sHTML []byte

// images are s.html's pictures, served at /s/<name>: the robots and the screenshots.
//
//go:embed s/*.png s/*.webp
var images embed.FS

// imageTypes are the images' content types by extension.
var imageTypes = map[string]string{".png": "image/png", ".webp": "image/webp"}

//go:embed app.html apps/*.html
var appFiles embed.FS

// app is one of s.i41.eu's apps: its page s.i41.eu/<ID> is app.html around apps/<ID>.html.
type app struct{ ID, Name string }

// apps are in the order of s.html's apps: the apps, then the apps for developers.
var apps = []app{{"pet", "Pet"}, {"focus", "Focus"}, {"raw", "Raw data"}, {"dev", "Dev"}}

// appPages are the apps' pages by ID.
var appPages = renderApps()

func renderApps() map[string][]byte {
	pages := map[string][]byte{}
	for _, a := range apps {
		t := template.Must(template.ParseFS(appFiles, "app.html", "apps/"+a.ID+".html"))
		var b bytes.Buffer
		if err := t.Execute(&b, struct {
			app
			Apps []app
		}{a, apps}); err != nil {
			panic(err)
		}
		pages[a.ID] = b.Bytes()
	}
	return pages
}

// isS tells whether the request is for s.i41.eu; any other host gets the i41.eu page.
func isS(host string) bool {
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	return strings.EqualFold(host, "s.i41.eu")
}

func writePage(w http.ResponseWriter, p []byte) {
	h := w.Header()
	h.Set("Content-Type", "text/html; charset=utf-8")
	h.Set("Cache-Control", "public, max-age=300")
	h.Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; img-src 'self' data:; base-uri 'none'; form-action 'none'; frame-ancestors 'none'")
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
	w.Write(p)
}

func handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		if isS(r.Host) {
			writePage(w, sHTML)
			return
		}
		writePage(w, indexHTML)
	})
	// an app's page, on s.i41.eu only
	mux.HandleFunc("GET /{app}", func(w http.ResponseWriter, r *http.Request) {
		p, ok := appPages[r.PathValue("app")]
		if !ok || !isS(r.Host) {
			http.NotFound(w, r)
			return
		}
		writePage(w, p)
	})
	// one picture by name, on s.i41.eu only: no listing, and {name} cannot hold a "/"
	mux.HandleFunc("GET /s/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		ct := imageTypes[path.Ext(name)]
		b, err := images.ReadFile("s/" + name)
		if !isS(r.Host) || ct == "" || err != nil {
			http.NotFound(w, r)
			return
		}
		h := w.Header()
		h.Set("Content-Type", ct)
		h.Set("Cache-Control", "public, max-age=86400")
		h.Set("X-Content-Type-Options", "nosniff")
		w.Write(b)
	})
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("ok\n"))
	})
	return mux
}

func main() {
	listen := flag.String("listen", ":8080", "address to listen on")
	flag.Parse()
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	srv := &http.Server{
		Addr:              *listen,
		Handler:           handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Info("i41eu-web listening", "listen", *listen)
	if err := srv.ListenAndServe(); err != nil {
		log.Error("listen", "err", err)
		os.Exit(1)
	}
}

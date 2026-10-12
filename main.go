// Command i41eu-web serves i41.eu: one page about the i41 home platform, with links to its
// parts (s.i41.eu, m.i41.eu, the apps on q41.eu) and to the code. The page is embedded.
package main

import (
	_ "embed"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"time"
)

//go:embed index.html
var indexHTML []byte

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
		writePage(w, indexHTML)
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

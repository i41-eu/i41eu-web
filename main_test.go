package main

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

func get(t *testing.T, path string) (int, string, http.Header) {
	t.Helper()
	rec := httptest.NewRecorder()
	handler().ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
	return rec.Code, rec.Body.String(), rec.Header()
}

func TestPage(t *testing.T) {
	code, body, h := get(t, "/")
	if code != 200 || !strings.Contains(body, "<h1>i41</h1>") {
		t.Fatalf("/: %d", code)
	}
	if !strings.Contains(h.Get("Content-Security-Policy"), "default-src 'none'") {
		t.Errorf("no CSP: %q", h.Get("Content-Security-Policy"))
	}
	for _, want := range []string{`href="https://s.i41.eu"`, `href="https://s.i41.eu/s3d/"`, `href="https://m.i41.eu"`, `href="https://github.com/i41-eu"`} {
		if !strings.Contains(body, want) {
			t.Errorf("no link %s", want)
		}
	}
	if code, body, _ := get(t, "/healthz"); code != 200 || body != "ok\n" {
		t.Errorf("/healthz: %d %q", code, body)
	}
	if code, _, _ := get(t, "/nope"); code != 404 {
		t.Errorf("/nope: %d, want 404", code)
	}
}

// Links to other sites open in a new tab; links to our own (i41.eu, q41.eu) stay in the tab.
func TestLinks(t *testing.T) {
	_, body, _ := get(t, "/")
	ours := regexp.MustCompile(`^https://([a-z0-9-]+\.)?(i41|q41)\.eu(/|$)`)
	for _, m := range regexp.MustCompile(`<a ([^>]*)>`).FindAllStringSubmatch(body, -1) {
		href := regexp.MustCompile(`href="([^"]*)"`).FindStringSubmatch(m[1])
		if href == nil {
			continue
		}
		newTab := strings.Contains(m[1], `target="_blank"`) && strings.Contains(m[1], `rel="noopener"`)
		if ours.MatchString(href[1]) == newTab {
			t.Errorf("%s: our sites stay in the tab, others open a new one (with rel=noopener)", href[1])
		}
	}
}

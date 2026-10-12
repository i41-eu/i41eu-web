package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path"
	"regexp"
	"strings"
	"testing"
)

// get serves one request for host and path.
func get(t *testing.T, host, p string) (int, string, http.Header) {
	t.Helper()
	req := httptest.NewRequest("GET", p, nil)
	req.Host = host
	rec := httptest.NewRecorder()
	handler().ServeHTTP(rec, req)
	return rec.Code, rec.Body.String(), rec.Header()
}

func TestPage(t *testing.T) {
	code, body, h := get(t, "i41.eu", "/")
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
	if code, body, _ := get(t, "i41.eu", "/healthz"); code != 200 || body != "ok\n" {
		t.Errorf("/healthz: %d %q", code, body)
	}
	if code, _, _ := get(t, "i41.eu", "/nope"); code != 404 {
		t.Errorf("/nope: %d, want 404", code)
	}
}

// s.i41.eu gets the robot's page; any other host the i41.eu page.
func TestHosts(t *testing.T) {
	for host, want := range map[string][]byte{
		"i41.eu":         indexHTML,
		"localhost:8080": indexHTML,
		"m.i41.eu":       indexHTML,
		"s.i41.eu.evil":  indexHTML,
		"s.i41.eu":       sHTML,
		"S.i41.eu:443":   sHTML,
	} {
		if _, body, _ := get(t, host, "/"); body != string(want) {
			t.Errorf("%s: wrong page", host)
		}
	}
}

func TestImages(t *testing.T) {
	for p, want := range map[string]int{
		"/s/robot3d-focus.png": 200,
		"/s/robot3d-dev.png":   200,
		"/s/":                  404,
		"/s/nope.webp":         404,
		"/s/%2e%2e%2fmain.go":  404,
		"/main.go":             404,
		"/x/main.go":           404,
	} {
		code, _, h := get(t, "s.i41.eu", p)
		if code != want {
			t.Errorf("%s: %d, want %d", p, code, want)
		}
		if want == 200 && h.Get("Content-Type") != imageTypes[path.Ext(p)] {
			t.Errorf("%s: %s", p, h.Get("Content-Type"))
		}
	}
	// the mux redirects to the cleaned path, which is not a picture
	if code, _, h := get(t, "s.i41.eu", "/s/../go.mod"); code == 200 || h.Get("Location") != "/go.mod" {
		t.Errorf("/s/../go.mod: %d to %q", code, h.Get("Location"))
	}
	if code, _, _ := get(t, "s.i41.eu", "/go.mod"); code != 404 {
		t.Errorf("/go.mod: %d, want 404", code)
	}
	if code, _, _ := get(t, "i41.eu", "/s/robot3d-dev.png"); code != 404 {
		t.Errorf("i41.eu/s/robot3d-dev.png: %d, want 404", code)
	}
	// every image is named by s.html, and every image a page names is there
	names, _ := images.ReadDir("s")
	for _, n := range names {
		if !bytes.Contains(sHTML, []byte("/s/"+n.Name())) {
			t.Errorf("s.html does not name %s", n.Name())
		}
	}
	for name, page := range allPages() {
		for _, m := range regexp.MustCompile(`"/s/([^"]+)"`).FindAllSubmatch(page, -1) {
			if _, err := images.ReadFile("s/" + string(m[1])); err != nil {
				t.Errorf("%s names a missing image: /s/%s", name, m[1])
			}
		}
	}
}

// ours are our own sites: i41.eu, q41.eu and their subdomains, mj41.cz.
var ours = regexp.MustCompile(`(^|\.)(i41\.eu|q41\.eu|mj41\.cz)$`)

// Links to other sites open in a new tab, so the visitor keeps this page; links to our own
// sites stay in the tab.
func TestExternalLinksOpenANewTab(t *testing.T) {
	link := regexp.MustCompile(`<a\b[^>]*\bhref="https?://([^/"]+)[^"]*"[^>]*>`)
	for name, page := range allPages() {
		for _, m := range link.FindAllSubmatch(page, -1) {
			tag, host := string(m[0]), string(m[1])
			newTab := bytes.Contains(m[0], []byte(`target="_blank"`)) && bytes.Contains(m[0], []byte(`rel="noopener"`))
			if ours.MatchString(host) == newTab {
				t.Errorf("%s: %s: our sites stay in the tab, others open a new one (with rel=noopener)", name, tag)
			}
			if ours.MatchString(host) && bytes.Contains(m[0], []byte(`data-ext="`)) {
				t.Errorf("%s: %s: a chip on a link to our own site", name, tag)
			}
		}
	}
}

// On s.i41.eu every link to another site is marked after it: a chip with the site ("GitHub",
// "m5stack.com"), unless the link's own text names it ("github.com/mj41", "infinite.pm").
func TestExternalLinksMarked(t *testing.T) {
	link := regexp.MustCompile(`(?s)<a\b([^>]*\bhref="https?://([^/"]+)[^"]*"[^>]*)>(.*?)</a>`)
	for name, page := range allPages() {
		if name == "index.html" {
			continue
		}
		for _, m := range link.FindAllSubmatch(page, -1) {
			attrs, host, text := string(m[1]), strings.ToLower(string(m[2])), strings.ToLower(string(m[3]))
			if ours.MatchString(host) || strings.Contains(attrs, `data-ext="`) {
				continue
			}
			site := strings.TrimPrefix(strings.TrimPrefix(host, "www."), "shop.")
			if strings.HasSuffix(host, "github.com") {
				site = "github"
			}
			if !strings.Contains(text, site) {
				t.Errorf("%s: %q to %s has no chip and does not say where it goes", name, text, host)
			}
		}
	}
}

// allPages are every page served, by name.
func allPages() map[string][]byte {
	pages := map[string][]byte{"index.html": indexHTML, "s.html": sHTML}
	for id, p := range appPages {
		pages["apps/"+id+".html"] = p
	}
	return pages
}

// Each app has a page on s.i41.eu only, and s.html's robots link to the apps' anchors, where
// each app links to its page.
func TestAppPages(t *testing.T) {
	s := string(sHTML)
	for _, a := range apps {
		code, body, _ := get(t, "s.i41.eu", "/"+a.ID)
		if code != 200 || !strings.Contains(body, "<h1>"+a.Name+"</h1>") {
			t.Errorf("s.i41.eu/%s: %d, not its page", a.ID, code)
		}
		// the apps s3d knows link to it
		if a.ID != "dev" && !strings.Contains(body, `href="https://s.i41.eu/s3d/#`+a.ID+`"`) {
			t.Errorf("s.i41.eu/%s: no link to s.i41.eu/s3d/#%s", a.ID, a.ID)
		}
		if !strings.Contains(body, `href="/`+a.ID+`" aria-current="page"`) {
			t.Errorf("s.i41.eu/%s: the app switcher does not mark it", a.ID)
		}
		if code, _, _ := get(t, "i41.eu", "/"+a.ID); code != 404 {
			t.Errorf("i41.eu/%s: %d, want 404", a.ID, code)
		}
		if a.ID != "dev" && !strings.Contains(s, `<a href="#`+a.ID+`">`) { // Dev has no robot picture yet
			t.Errorf("s.html: no robot picture links to #%s", a.ID)
		}
		for _, want := range []string{`id="` + a.ID + `"`, `<h3><a href="/` + a.ID + `">`} {
			if !strings.Contains(s, want) {
				t.Errorf("s.html: no %s", want)
			}
		}
	}
	if code, _, _ := get(t, "s.i41.eu", "/nope"); code != 404 {
		t.Errorf("s.i41.eu/nope: %d, want 404", code)
	}
	// the robots and the apps in the same order
	last := 0
	for _, a := range apps {
		if i := strings.Index(s, `id="`+a.ID+`"`); i < last {
			t.Errorf("s.html: %s out of order", a.ID)
		} else {
			last = i
		}
	}
}

// The robot's QR picture leads to the section on how a robot connects: USB, its manager, the
// QR code and the sign-in.
func TestConnectSection(t *testing.T) {
	s := string(sHTML)
	if !strings.Contains(s, `<a href="#connect"><figure><img src="/s/robot3d-qr.png"`) {
		t.Errorf("s.html: the QR picture does not link to #connect")
	}
	i := strings.Index(s, `<section class="card" id="connect">`)
	if i < 0 {
		t.Fatalf("s.html: no #connect section")
	}
	sec := s[i : i+strings.Index(s[i:], "</section>")]
	for _, want := range []string{"USB", "The robot alone", "Your home manager", "m.i41.eu", "QR", "GitHub or Google"} {
		if !strings.Contains(sec, want) {
			t.Errorf("#connect: no %q", want)
		}
	}
}

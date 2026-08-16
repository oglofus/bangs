package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBuildRedirectURLEncodesQuery(t *testing.T) {
	redirectURL, ok := buildRedirectURL([]byte("https://example.com/search?q="+string([]byte{QueryPlaceholder})), "red & blue")
	if !ok {
		t.Fatal("expected a valid redirect URL")
	}

	if want := "https://example.com/search?q=red+%26+blue"; redirectURL != want {
		t.Fatalf("redirect URL = %q, want %q", redirectURL, want)
	}
}

func TestQueryHandlerIgnoresArbitraryFallback(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?q=example&fallback=https://attacker.example/", nil)
	res := httptest.NewRecorder()

	queryHandler(res, req)

	if res.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusFound)
	}
	if got := res.Header().Get("Location"); got != "https://www.google.com/search?q=example" {
		t.Fatalf("Location = %q, want the fixed Google fallback", got)
	}
}

func TestQueryHandlerRedirectsKnownBang(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?q=example+!w", nil)
	res := httptest.NewRecorder()

	queryHandler(res, req)

	if res.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusFound)
	}
	if got := res.Header().Get("Location"); got != "https://en.wikipedia.org/wiki/Special:Search?search=example" {
		t.Fatalf("Location = %q, want the Wikipedia bang redirect", got)
	}
}

func TestQueryHandlerRendersHostAwareSEO(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://customer.example/", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	res := httptest.NewRecorder()

	queryHandler(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	body := res.Body.String()
	for _, expected := range []string{
		`<meta content="https://customer.example/" property="og:url" />`,
		`<link href="https://customer.example/" rel="canonical" />`,
		`<meta content="https://customer.example/" name="twitter:url" />`,
		`https://bangs.oglofus.com/#website`,
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("HTML does not contain %q", expected)
		}
	}

	start := strings.Index(body, `<script type="application/ld+json">`)
	if start < 0 {
		t.Fatal("JSON-LD script was not rendered")
	}
	end := strings.Index(body[start:], `</script>`)
	if end < 0 {
		t.Fatal("JSON-LD script was not closed")
	}
	jsonLD := body[start+len(`<script type="application/ld+json">`) : start+end]
	if !json.Valid([]byte(jsonLD)) {
		t.Fatalf("JSON-LD is invalid JSON: %s", jsonLD)
	}
}

func TestSEOEndpointsUseRequestOrigin(t *testing.T) {
	tests := []struct {
		path        string
		contentType string
		contains    string
	}{
		{"/robots.txt", "text/plain; charset=utf-8", "Sitemap: https://customer.example/sitemap.xml"},
		{"/sitemap.xml", "application/xml; charset=utf-8", "<loc>https://customer.example/</loc>"},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "https://customer.example"+test.path, nil)
			res := httptest.NewRecorder()

			queryHandler(res, req)

			if got := res.Header().Get("Content-Type"); got != test.contentType {
				t.Fatalf("Content-Type = %q, want %q", got, test.contentType)
			}
			if !strings.Contains(res.Body.String(), test.contains) {
				t.Fatalf("response does not contain %q", test.contains)
			}
		})
	}
}

func TestRequestOriginFallsBackForInvalidHost(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://customer.example/", nil)
	req.Host = "attacker.example/<script>"

	if got := requestOrigin(req); got != canonicalOrigin {
		t.Fatalf("requestOrigin = %q, want %q", got, canonicalOrigin)
	}
}

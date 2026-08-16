package main

import (
	"net/http"
	"net/http/httptest"
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

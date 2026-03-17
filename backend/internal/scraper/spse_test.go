package scraper

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExtractToken(t *testing.T) {
	html := `<html><body><script>var authenticityToken = 'abc123xyz';</script></body></html>`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(html))
	}))
	defer ts.Close()

	client := NewSpseClient(ts.URL, 2026)
	token, err := client.extractToken("/")
	if err != nil {
		t.Fatalf("extractToken: %v", err)
	}
	if token != "abc123xyz" {
		t.Errorf("expected 'abc123xyz', got %q", token)
	}
}

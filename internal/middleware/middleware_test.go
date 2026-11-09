package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChainRequestIDAndRecover(t *testing.T) {
	var seen string
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = GetRequestID(r.Context())
		panic("boom")
	}), RequestID, Logger, Recover)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if seen == "" || rec.Header().Get(RequestIDHeader) != seen {
		t.Fatalf("request id: ctx=%q header=%q", seen, rec.Header().Get(RequestIDHeader))
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(RequestIDHeader, "abc")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if seen != "abc" {
		t.Fatalf("incoming request id not reused: %q", seen)
	}
}

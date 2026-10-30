package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestVersionIsServedBothWays(t *testing.T) {
	router := testRouter(t)
	for _, path := range []string{"/version", "/api/version"} {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"version":"`) {
			t.Errorf("%s = %d %s", path, rec.Code, rec.Body.String())
		}
	}
}

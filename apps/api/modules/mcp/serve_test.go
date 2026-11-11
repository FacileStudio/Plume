package mcp_test

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"testing"

	"github.com/FacileStudio/porte"
	"github.com/go-pdf/fpdf"
)

// envServe is the address TestServe listens on, empty to skip it.
func envServe() string {
	return os.Getenv("MCP_SERVE")
}

// waitInterrupt blocks until the process is interrupted.
func waitInterrupt() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
}

// itoa renders an id.
func itoa(n int64) string {
	return strconv.FormatInt(n, 10)
}

// samplePDF is a PDF of pages blank pages.
func samplePDF(t testing.TB, pages int) []byte {
	t.Helper()
	pdf := fpdf.New("P", "mm", "A4", "")
	for range pages {
		pdf.AddPage()
	}
	var out bytes.Buffer
	must(t, pdf.Output(&out))
	return out.Bytes()
}

// encoded is samplePDF in base64.
func encoded(t testing.TB, pages int) string {
	return base64.StdEncoding.EncodeToString(samplePDF(t, pages))
}

// requireTestSession stands in for porte on the confirm page: X-Test-User is the
// signed-in user, and X-Test-Actor makes the request a delegated token.
func requireTestSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := strconv.ParseInt(r.Header.Get("X-Test-User"), 10, 64)
		if err != nil {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}
		identity := porte.Identity{UserID: userID, SessionID: 1, Actor: r.Header.Get("X-Test-Actor")}
		if identity.Actor != "" {
			identity.SessionID = 0
		}
		next.ServeHTTP(w, r.WithContext(porte.WithIdentity(r.Context(), identity)))
	})
}

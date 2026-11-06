package nuage

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"unicode/utf8"
)

func quietLogger() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

type fakeTokens struct {
	err   error
	users atomic.Int64
}

func (f *fakeTokens) Token(_ context.Context, userID int64) (string, error) {
	f.users.Store(userID)
	return "delegated", f.err
}

func fastPublisher(url string, tokens TokenSource) *Publisher {
	p := NewPublisher(NewClient(url), tokens, quietLogger())
	p.backoff = time.Millisecond
	return p
}

func loadPDF() ([]byte, error) { return []byte("%PDF-1.7"), nil }

func TestPutSendsKeyNameAndTokenAndAcceptsReplace(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusCreated} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			if r.Method != http.MethodPut || r.URL.Path != "/api/app/object" {
				t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			}
			if r.URL.Query().Get("key") != "plume:document:3" || r.URL.Query().Get("name") != "a b.pdf" {
				t.Errorf("bad query %v", r.URL.Query())
			}
			if r.Header.Get("Authorization") != "Bearer tok" || r.Header.Get("Content-Type") != "application/pdf" || string(body) != "x" {
				t.Errorf("bad headers or body %v %q", r.Header, body)
			}
			w.WriteHeader(status)
		}))
		err := NewClient(srv.URL).Put(context.Background(), "tok", "plume:document:3", "a b.pdf", []byte("x"))
		srv.Close()
		if err != nil {
			t.Fatalf("status %d: %v", status, err)
		}
	}
}

func TestPublishActsForTheOwnerAndRetriesOn5xx(t *testing.T) {
	var puts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if puts.Add(1) < 3 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	tokens := &fakeTokens{}
	err := fastPublisher(srv.URL, tokens).publishWithRetry(context.Background(), Job{DocumentID: 1, OwnerID: 42, Title: "x", Load: loadPDF}, quietLogger())
	if err != nil || puts.Load() != 3 || tokens.users.Load() != 42 {
		t.Fatalf("err=%v puts=%d owner=%d", err, puts.Load(), tokens.users.Load())
	}
}

func TestPublishGivesUpAfterMaxAttempts(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	err := fastPublisher(srv.URL, &fakeTokens{}).publishWithRetry(context.Background(), Job{DocumentID: 1, Load: loadPDF}, quietLogger())
	if err == nil || calls.Load() != maxAttempts {
		t.Fatalf("err=%v calls=%d", err, calls.Load())
	}
}

func TestPublishDoesNotRetryOn4xx(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusPaymentRequired)
	}))
	defer srv.Close()

	err := fastPublisher(srv.URL, &fakeTokens{}).publishWithRetry(context.Background(), Job{DocumentID: 1, Load: loadPDF}, quietLogger())
	var se *StatusError
	if !errors.As(err, &se) || se.Status != http.StatusPaymentRequired || calls.Load() != 1 {
		t.Fatalf("err=%v calls=%d", err, calls.Load())
	}
}

func TestPublishWithoutARegistreSessionIsNotRetriedAndNeverCallsNuage(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1) }))
	defer srv.Close()

	tokens := &fakeTokens{err: ErrNotConnected}
	err := fastPublisher(srv.URL, tokens).publishWithRetry(context.Background(), Job{DocumentID: 1, Load: loadPDF}, quietLogger())
	if !errors.Is(err, ErrNotConnected) || calls.Load() != 0 {
		t.Fatalf("err=%v nuage calls=%d", err, calls.Load())
	}
}

func TestEnqueueReportsTheStoredKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	done := make(chan string, 1)
	fastPublisher(srv.URL, &fakeTokens{}).Enqueue(Job{
		DocumentID: 1, OwnerID: 1, Title: "t", LinkedTo: "ardoise:contract:42:signed",
		Load: loadPDF, OnPublished: func(key string) { done <- key },
	})
	select {
	case key := <-done:
		if key != "ardoise:contract:42:signed" {
			t.Fatalf("OnPublished got %q", key)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("never published")
	}
}

func TestEnqueueNeverBlocksNorFailsWhenNuageIsDown(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-release
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	defer close(release)

	p := fastPublisher(srv.URL, &fakeTokens{})
	start := time.Now()
	for i := 0; i < 20; i++ {
		p.Enqueue(Job{DocumentID: int64(i), Load: loadPDF})
	}
	var nilPublisher *Publisher
	nilPublisher.Enqueue(Job{})
	if elapsed := time.Since(start); elapsed > 100*time.Millisecond {
		t.Fatalf("Enqueue blocked for %s", elapsed)
	}
}

func TestObjectKey(t *testing.T) {
	if got := objectKey(Job{DocumentID: 9}); got != "plume:document:9" {
		t.Errorf("default key = %q", got)
	}
	if got := objectKey(Job{DocumentID: 9, LinkedTo: "ardoise:1"}); got != "ardoise:1" {
		t.Errorf("explicit key = %q", got)
	}
}

func TestFilename(t *testing.T) {
	cases := map[string]string{
		"Contract":             "Contract.pdf",
		"a/b.PDF":              "a-b.PDF",
		"  ":                   "document.pdf",
		"2026/001":             "2026-001.pdf",
		"a\x00b\u202ec\u200bd": "abcd.pdf",
	}
	for in, want := range cases {
		if got := filename(in); got != want {
			t.Errorf("filename(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFilenameIsCappedAtNuageLimitWithoutSplittingACharacter(t *testing.T) {
	for _, title := range []string{strings.Repeat("é", 300), strings.Repeat("😀", 100), strings.Repeat("a", 400) + ".pdf"} {
		name := filename(title)
		if len(name) > 255 || !utf8.ValidString(name) || !strings.HasSuffix(name, ".pdf") {
			t.Errorf("filename = %d bytes valid=%v suffix=%v", len(name), utf8.ValidString(name), strings.HasSuffix(name, ".pdf"))
		}
	}
}

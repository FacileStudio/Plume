package nuage

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func quietLogger() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func fastPublisher(url, defaultSpace string) *Publisher {
	p := NewPublisher(NewClient(url, "tok"), defaultSpace, "Plume", quietLogger())
	p.backoff = time.Millisecond
	return p
}

func loadPDF() ([]byte, error) { return []byte("%PDF-1.7"), nil }

func TestEnsureSpaceMemoisesAndHandlesBothStatuses(t *testing.T) {
	var puts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		puts.Add(1)
		if r.Header.Get("Authorization") != "Bearer tok" || r.URL.Path != "/api/spaces/by-facile-id/fac_a" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(map[bool]int{true: 201, false: 200}[puts.Load() == 1])
		_, _ = w.Write([]byte(`{"id":7}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "tok")
	for i := 0; i < 3; i++ {
		id, err := c.EnsureSpace(context.Background(), "fac_a", "A")
		if err != nil || id != 7 {
			t.Fatalf("EnsureSpace = %d, %v", id, err)
		}
	}
	if puts.Load() != 1 {
		t.Fatalf("expected one PUT, got %d", puts.Load())
	}
}

func TestUploadAcceptsExistingFile(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusCreated} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseMultipartForm(1 << 20); err != nil {
				t.Errorf("parse form: %v", err)
			}
			if r.FormValue("space_id") != "7" || r.FormValue("linked_to") != "plume:document:3" {
				t.Errorf("bad form: %v", r.MultipartForm.Value)
			}
			if _, _, err := r.FormFile("file"); err != nil {
				t.Errorf("missing file: %v", err)
			}
			w.WriteHeader(status)
		}))
		err := NewClient(srv.URL, "tok").Upload(context.Background(), 7, "a.pdf", "plume:document:3", []byte("x"))
		srv.Close()
		if err != nil {
			t.Fatalf("status %d: %v", status, err)
		}
	}
}

func TestPublishRetriesOn5xx(t *testing.T) {
	var uploads atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			_, _ = w.Write([]byte(`{"id":1}`))
			return
		}
		if uploads.Add(1) < 3 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	p := fastPublisher(srv.URL, "fac_d")
	err := p.publishWithRetry(context.Background(), Job{DocumentID: 1, Title: "x", Load: loadPDF}, quietLogger())
	if err != nil || uploads.Load() != 3 {
		t.Fatalf("err=%v uploads=%d", err, uploads.Load())
	}
}

func TestPublishGivesUpAfterMaxAttempts(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	p := fastPublisher(srv.URL, "fac_d")
	err := p.publishWithRetry(context.Background(), Job{DocumentID: 1, Load: loadPDF}, quietLogger())
	if err == nil || calls.Load() != maxAttempts {
		t.Fatalf("err=%v calls=%d", err, calls.Load())
	}
}

func TestPublishDoesNotRetryOn4xx(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	p := fastPublisher(srv.URL, "fac_d")
	err := p.publishWithRetry(context.Background(), Job{DocumentID: 1, Load: loadPDF}, quietLogger())
	var se *StatusError
	if !errors.As(err, &se) || se.Status != 403 || calls.Load() != 1 {
		t.Fatalf("err=%v calls=%d", err, calls.Load())
	}
}

func TestPublishWithoutAnySpaceIsNotRetried(t *testing.T) {
	p := fastPublisher("http://127.0.0.1:1", "")
	err := p.publishWithRetry(context.Background(), Job{DocumentID: 1, Load: loadPDF}, quietLogger())
	if !errors.Is(err, errNoSpace) {
		t.Fatalf("err=%v", err)
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

	p := fastPublisher(srv.URL, "fac_d")
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

func TestFilename(t *testing.T) {
	cases := map[string]string{"Contract": "Contract.pdf", "a/b.PDF": "a-b.PDF", "  ": "document.pdf"}
	for in, want := range cases {
		if got := filename(in); got != want {
			t.Errorf("filename(%q) = %q, want %q", in, got, want)
		}
	}
}

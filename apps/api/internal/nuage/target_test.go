package nuage

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestResolveTarget(t *testing.T) {
	cases := []struct {
		name string
		job  Job
		want Target
		err  bool
	}{
		{"defaults", Job{DocumentID: 9}, Target{"fac_default", "Plume", "plume:document:9"}, false},
		{"plume space", Job{DocumentID: 9, SpaceFacileID: "fac_s", SpaceName: "S"}, Target{"fac_s", "S", "plume:document:9"}, false},
		{"explicit key", Job{DocumentID: 9, LinkedTo: "ardoise:1"}, Target{"fac_default", "Plume", "ardoise:1"}, false},
		{"explicit space and key", Job{DocumentID: 9, SpaceFacileID: "fac_x", LinkedTo: "ardoise:1"}, Target{"fac_x", "Plume", "ardoise:1"}, false},
	}
	for _, c := range cases {
		got, err := resolveTarget(c.job, "fac_default", "Plume")
		if err != nil || got != c.want {
			t.Errorf("%s: got %+v, %v want %+v", c.name, got, err, c.want)
		}
	}
	if _, err := resolveTarget(Job{DocumentID: 9}, "", "Plume"); err == nil {
		t.Error("no space anywhere must fail")
	}
}

func TestPublishUsesTheRequestedTargetAndReportsIt(t *testing.T) {
	var mu sync.Mutex
	var paths []string
	var linked string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		paths = append(paths, r.URL.Path)
		if r.Method == http.MethodPost {
			_ = r.ParseMultipartForm(1 << 20)
			linked = r.FormValue("linked_to")
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":4}`))
	}))
	defer srv.Close()

	done := make(chan [2]string, 1)
	fastPublisher(srv.URL, "").Enqueue(Job{
		DocumentID: 1, Title: "t", SpaceFacileID: "fac_0123456789abcdef01234567", LinkedTo: "ardoise:contract:42:signed",
		Load:        loadPDF,
		OnPublished: func(space, key string) { done <- [2]string{space, key} },
	})
	select {
	case got := <-done:
		if got != [2]string{"fac_0123456789abcdef01234567", "ardoise:contract:42:signed"} {
			t.Fatalf("OnPublished got %v", got)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("never published")
	}
	mu.Lock()
	defer mu.Unlock()
	if linked != "ardoise:contract:42:signed" || paths[0] != "/api/spaces/by-facile-id/fac_0123456789abcdef01234567" {
		t.Fatalf("requests %v linked_to %q", paths, linked)
	}
}

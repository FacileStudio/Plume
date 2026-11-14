package reminders

import (
	"context"
	"io"
	"log/slog"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/FacileStudio/Plume/apps/api/schemas"
	"github.com/FacileStudio/tronc/testdb"

	"gorm.io/gorm"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	url, ok := testdb.URL()
	if !ok {
		testdb.Announce("sh scripts/postgres-up.sh")
		os.Exit(m.Run())
	}
	db, err := testdb.Open(url, testdb.Config{Prefix: "plume_reminders_test", Migrate: schemas.Migrate})
	if err != nil {
		panic(err)
	}
	testDB = db
	os.Exit(m.Run())
}

func countingService(runs *atomic.Int32) *Service {
	s := NewService(testDB, nil, nil, "")
	s.tick = 20 * time.Millisecond
	s.run = func(context.Context) (int, error) {
		runs.Add(1)
		return 0, nil
	}
	return s
}

func TestOnlyTheLockHolderTicks(t *testing.T) {
	if testDB == nil {
		t.Skip(testdb.SkipReason("sh scripts/postgres-up.sh"))
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	var runsA, runsB atomic.Int32
	ctxA, cancelA := context.WithCancel(t.Context())
	defer cancelA()
	Start(ctxA, countingService(&runsA), logger)
	Start(t.Context(), countingService(&runsB), logger)

	waitFor(t, func() bool { return runsA.Load()+runsB.Load() >= 3 })
	first, second := &runsA, &runsB
	if runsA.Load() == 0 {
		first, second = &runsB, &runsA
	}
	time.Sleep(200 * time.Millisecond)
	if got := second.Load(); got != 0 {
		t.Fatalf("the replica without the lock ticked %d times", got)
	}
	if first.Load() < 3 {
		t.Fatalf("the holder ticked %d times, want at least 3", first.Load())
	}
	if first == &runsA {
		cancelA()
		waitFor(t, func() bool { return runsB.Load() >= 2 })
	}
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("timed out")
}

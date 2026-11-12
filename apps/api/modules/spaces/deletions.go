package spaces

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/FacileStudio/porte/directory"
	"github.com/FacileStudio/porte/kit"

	"gorm.io/gorm"
)

// Deletions erases the people Registre erased. It polls Registre's deletion
// list with the same client credentials as the snapshot sync and accepts
// Registre's signed push on directory.DeletionsPath.
type Deletions struct {
	users     *kit.Store
	freeze    func(ctx context.Context, docID int64) error
	uploadDir string
	receiver  *directory.Deletions
}

// DeletionsConfig is what NewDeletions builds the receiver from. Users is the
// kit store whose EraseAccount deletes the account; Freeze writes a final
// document's audit trail before the owner's name and email are erased; Erase
// is porte's own rows eraser, pg.Store.EraseIdentity; Holds is
// pg.Store.HoldsIdentity, so a handled subject that an account holds again is
// erased again on the next push or poll.
type DeletionsConfig struct {
	ORM       *gorm.DB
	Client    directory.Client
	Users     *kit.Store
	Freeze    func(ctx context.Context, docID int64) error
	Erase     func(context.Context, string, string) error
	Holds     func(context.Context, string, string) (bool, error)
	UploadDir string
	Logger    *slog.Logger
}

// NewDeletions builds the receiver over Plume's tables.
func NewDeletions(cfg DeletionsConfig) *Deletions {
	d := &Deletions{users: cfg.Users, freeze: cfg.Freeze, uploadDir: cfg.UploadDir}
	d.receiver = &directory.Deletions{
		Client:          cfg.Client,
		OnPersonDeleted: d.ErasePerson,
		Erase:           cfg.Erase,
		Holds:           cfg.Holds,
		State:           deletionState{orm: cfg.ORM},
		Logger:          cfg.Logger,
	}
	return d
}

// ErasedLedger answers whether Registre erased a subject Plume handled, from
// the state the receiver saves, for oidc.Deps.Erased and the snapshot sync. It
// never runs the receiver and reads the state on every call, so it needs only
// the issuer and answers the same on any replica.
func ErasedLedger(orm *gorm.DB, issuer string) *directory.Deletions {
	return &directory.Deletions{Client: directory.Client{Issuer: issuer}, State: deletionState{orm: orm}}
}

// Run polls until ctx ends. Start it in exactly one replica.
func (d *Deletions) Run(ctx context.Context) {
	d.receiver.Run(ctx)
}

// Handler verifies and applies Registre's signed user.deleted push.
func (d *Deletions) Handler() http.Handler {
	return d.receiver.Handler()
}

type deletionState struct {
	orm *gorm.DB
}

// Load returns the saved state, or nil before the first save.
func (s deletionState) Load(ctx context.Context) ([]byte, error) {
	var rows []struct{ State []byte }
	err := s.orm.WithContext(ctx).Raw(`SELECT state FROM directory_deletion_state WHERE id = 1`).Scan(&rows).Error
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	return rows[0].State, nil
}

// Save replaces the saved state.
func (s deletionState) Save(ctx context.Context, state []byte) error {
	return s.orm.WithContext(ctx).Exec(
		`INSERT INTO directory_deletion_state (id, state, updated_at) VALUES (1, ?, now())
		 ON CONFLICT (id) DO UPDATE SET state = EXCLUDED.state, updated_at = now()`, state).Error
}

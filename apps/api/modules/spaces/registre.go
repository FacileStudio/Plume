package spaces

import (
	"context"
	"log/slog"
	"time"

	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/directory"

	"gorm.io/gorm"
)

// maxImportRetry caps the gap between two import attempts, so a Registre
// that is not deployed yet costs one request every few minutes.
const maxImportRetry = 5 * time.Minute

// Registre keeps Plume's spaces equal to the ones Registre owns. issuer is
// the configured OIDC_ISSUER: it is the provider column of every porte
// identity a snapshot member is matched on. erased, when set, is the ledger
// of the subjects Registre erased: a snapshot member it names is skipped, so a
// stale snapshot cannot give an erased person an account again.
type Registre struct {
	orm    *gorm.DB
	issuer string
	erased porte.ErasedSubjects
	logger *slog.Logger
	syncer *directory.Syncer
}

// NewRegistre wires the Registre link to its database, to the directory
// client, which carries the issuer and the app's own OIDC client credentials,
// and to the erased-subjects ledger, which may be nil.
func NewRegistre(orm *gorm.DB, client directory.Client, erased porte.ErasedSubjects, logger *slog.Logger) *Registre {
	r := &Registre{orm: orm, issuer: client.Issuer, erased: erased, logger: logger}
	r.syncer = &directory.Syncer{Client: client, Apply: r.Apply, Logger: logger}
	return r
}

// Run hands Registre every space Plume held before Registre owned spaces,
// then runs the syncer until ctx ends.
//
// The import comes first and is retried, doubling from first up to
// maxImportRetry, until Registre accepts it: a snapshot applied before the
// import would remove every member Registre has not heard of yet. Only the
// first failure is a warning, so a Registre answering 404 for a week leaves
// one line in the log rather than two thousand.
func (r *Registre) Run(ctx context.Context, first time.Duration) {
	for attempt, delay := 0, first; ; attempt, delay = attempt+1, min(delay*2, maxImportRetry) {
		err := r.importOnce(ctx)
		if err == nil {
			break
		}
		level := slog.LevelDebug
		if attempt == 0 {
			level = slog.LevelWarn
		}
		r.logger.Log(ctx, level, "registre: space import failed, retrying",
			slog.Any("error", err), slog.Duration("retry_in", delay))
		if !wait(ctx, delay) {
			return
		}
	}
	r.logger.Info("registre: spaces imported, syncing")
	r.syncer.Run(ctx)
}

// importOnce sends the unlinked spaces to Registre once.
func (r *Registre) importOnce(ctx context.Context) error {
	spaces, err := r.ImportSpaces(ctx)
	if err != nil {
		return err
	}
	_, err = r.syncer.Client.Import(ctx, spaces)
	return err
}

// importRow is one membership of an unlinked space, as ImportSpaces reads it.
type importRow struct {
	SpaceID int64
	Name    string
	Subject string
	Email   string
	Role    string
}

// ImportSpaces lists every space not yet linked to Registre with its members.
// A member signed in through the configured issuer is sent by subject; any
// other member, a local password account included, by email.
func (r *Registre) ImportSpaces(ctx context.Context) ([]directory.ImportSpace, error) {
	var rows []importRow
	if err := r.orm.WithContext(ctx).Raw(importQuery, r.issuer).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := []directory.ImportSpace{}
	var last int64
	for _, row := range rows {
		if row.SpaceID != last {
			out = append(out, directory.ImportSpace{Name: row.Name})
			last = row.SpaceID
		}
		member := directory.ImportMember{Subject: row.Subject, Role: row.Role}
		if row.Subject == "" {
			member.Email = row.Email
		}
		out[len(out)-1].Members = append(out[len(out)-1].Members, member)
	}
	return out, nil
}

const importQuery = `
SELECT DISTINCT ON (s.id, m.id)
       s.id AS space_id, s.name, coalesce(i.subject, '') AS subject, u.email, m.role
  FROM spaces s
  JOIN space_members m ON m.space_id = s.id
  JOIN users u ON u.id = m.user_id
  LEFT JOIN porte_identities i ON i.user_id = u.id AND i.provider = ?
 WHERE s.registre_space_id IS NULL
 ORDER BY s.id, m.id
`

// wait sleeps for d or until ctx ends, reporting whether it slept fully.
func wait(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

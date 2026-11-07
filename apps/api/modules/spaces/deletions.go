package spaces

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/FacileStudio/porte/directory"

	"gorm.io/gorm"
)

// Deletions erases the people Registre erased. It polls Registre's deletion
// list with the same client credentials as the snapshot sync and accepts
// Registre's signed push on directory.DeletionsPath.
type Deletions struct {
	orm       *gorm.DB
	uploadDir string
	receiver  *directory.Deletions
}

// NewDeletions builds the receiver over Plume's tables. erase is porte's own
// rows eraser, pg.Store.EraseIdentity.
func NewDeletions(
	orm *gorm.DB, client directory.Client, erase func(context.Context, string, string) error,
	uploadDir string, logger *slog.Logger,
) *Deletions {
	d := &Deletions{orm: orm, uploadDir: uploadDir}
	d.receiver = &directory.Deletions{
		Client:          client,
		OnPersonDeleted: d.ErasePerson,
		Erase:           erase,
		State:           deletionState{orm: orm},
		Logger:          logger,
	}
	return d
}

// Run polls until ctx ends. Start it in exactly one replica.
func (d *Deletions) Run(ctx context.Context) {
	d.receiver.Run(ctx)
}

// Handler verifies and applies Registre's signed user.deleted push.
func (d *Deletions) Handler() http.Handler {
	return d.receiver.Handler()
}

const (
	soleOwnedSpaces = `SELECT s.id FROM spaces s
		WHERE s.owner_id = @id AND s.registre_space_id IS NULL
		AND NOT EXISTS (SELECT 1 FROM space_members m WHERE m.space_id = s.id AND m.user_id <> @id)`
	erasedDocuments = `SELECT id, storage_path FROM documents
		WHERE (owner_id = @id AND space_id IS NULL) OR space_id IN (` + soleOwnedSpaces + `)`
	deleteFields  = `DELETE FROM fields WHERE document_id IN @docs`
	deleteSigners = `DELETE FROM signers WHERE document_id IN @docs`
	deleteDocs    = `DELETE FROM documents WHERE id IN @docs`
)

const successor = `(SELECT m.user_id FROM space_members m WHERE m.space_id = %s AND m.user_id <> @id
		ORDER BY (m.role IN ('owner', 'admin')) DESC, m.joined_at, m.id LIMIT 1)`

var (
	transferDocuments = `UPDATE documents d SET owner_id = COALESCE(` + fmt.Sprintf(successor, "d.space_id") + `, 0),
		external_ref = CASE WHEN d.external_ref <> '' AND EXISTS (
			SELECT 1 FROM documents x WHERE x.external_ref = d.external_ref
			AND x.owner_id = COALESCE(` + fmt.Sprintf(successor, "d.space_id") + `, 0)) THEN '' ELSE d.external_ref END
		WHERE d.owner_id = @id AND d.space_id IS NOT NULL`
	transferSpaces = `UPDATE spaces s SET owner_id = COALESCE(` + fmt.Sprintf(successor, "s.id") + `, 0) WHERE s.owner_id = @id`
)

var accountStatements = []string{
	`DELETE FROM space_members WHERE space_id IN (` + soleOwnedSpaces + `)`,
	`DELETE FROM spaces WHERE id IN (` + soleOwnedSpaces + `)`,
	`UPDATE documents SET client_id = NULL WHERE client_id IN (SELECT id FROM clients WHERE owner_id = @id)`,
	`DELETE FROM clients WHERE owner_id = @id`,
	`DELETE FROM webhooks WHERE owner_id = @id`,
	`DELETE FROM smtp_configs WHERE owner_id = @id`,
	transferDocuments,
	transferSpaces,
	`DELETE FROM space_members WHERE user_id = @id`,
	`DELETE FROM users WHERE id = @id`,
}

type erasedDocument struct {
	ID          int64
	StoragePath string
}

// ErasePerson deletes the account behind a Registre identity and what it owns:
// its personal documents with their signers, fields and files, its contacts,
// webhooks, SMTP settings and memberships, and the spaces only it belonged to.
// Documents it created in a space others use stay and pass, like the space
// itself, to the space's longest-standing admin, else its longest-standing
// member, so they stay editable and their Nuage archiving keeps a live owner;
// with nobody left they fall to owner 0. A subject
// that is already gone is a no-op, so a replayed notice changes nothing.
func (d *Deletions) ErasePerson(ctx context.Context, provider, subject string) error {
	var account struct{ ID int64 }
	err := d.orm.WithContext(ctx).Raw(
		`SELECT u.id FROM users u JOIN porte_identities i ON i.user_id = u.id
		 WHERE i.provider = ? AND i.subject = ?`, provider, subject).Scan(&account).Error
	if err != nil || account.ID == 0 {
		return err
	}
	args := map[string]any{"id": account.ID}
	var docs []erasedDocument
	err = d.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Raw(erasedDocuments, args).Scan(&docs).Error; err != nil {
			return err
		}
		if len(docs) > 0 {
			docArgs := map[string]any{"docs": documentIDs(docs)}
			for _, statement := range []string{deleteFields, deleteSigners} {
				if err := tx.Exec(statement, docArgs).Error; err != nil {
					return err
				}
			}
			if err := tx.Exec(deleteDocs, docArgs).Error; err != nil {
				return err
			}
		}
		for _, statement := range accountStatements {
			if err := tx.Exec(statement, args).Error; err != nil {
				return fmt.Errorf("erase account %d: %w", account.ID, err)
			}
		}
		return nil
	})
	if err == nil {
		d.removeFiles(account.ID, docs)
	}
	return err
}

func documentIDs(docs []erasedDocument) []int64 {
	ids := make([]int64, len(docs))
	for i, doc := range docs {
		ids[i] = doc.ID
	}
	return ids
}

func (d *Deletions) removeFiles(ownerID int64, docs []erasedDocument) {
	for _, doc := range docs {
		var paths []string
		if doc.StoragePath != "" {
			original := filepath.Join(d.uploadDir, filepath.Clean("/"+doc.StoragePath))
			paths = append(paths, original, strings.TrimSuffix(original, ".pdf")+"_signed.pdf")
		}
		paths = append(paths,
			filepath.Join(d.uploadDir, "audit", fmt.Sprintf("audit_%d.pdf", doc.ID)),
			filepath.Join(d.uploadDir, "certificates", fmt.Sprintf("cert_%d.pdf", doc.ID)))
		for _, path := range paths {
			removeFile(path)
		}
	}
	_ = os.Remove(filepath.Join(d.uploadDir, fmt.Sprint(ownerID)))
}

func removeFile(path string) {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		slog.Warn("failed to remove an erased person's file", slog.Any("error", err))
	}
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

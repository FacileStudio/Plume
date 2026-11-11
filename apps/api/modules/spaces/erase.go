package spaces

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	mcpkit "github.com/FacileStudio/tronc/mcp"
)

const (
	accountOf = `SELECT i.user_id FROM porte_identities i
		WHERE i.provider IN ($1, $2) AND i.subject = $3`
	finalDocuments = `SELECT id FROM documents
		WHERE owner_id IN (` + accountOf + `) AND status IN ('completed', 'declined') ORDER BY id`
	soleOwnedSpaces = `SELECT s.id FROM spaces s
		WHERE s.owner_id = $1 AND s.registre_space_id IS NULL
		AND NOT EXISTS (SELECT 1 FROM space_members m WHERE m.space_id = s.id AND m.user_id <> $1)`
	erasedDrafts = `SELECT id FROM documents WHERE owner_id = $1 AND status = 'draft'
		AND (space_id IS NULL OR space_id IN (` + soleOwnedSpaces + `))`
	deleteDrafts = `DELETE FROM documents WHERE id IN (` + erasedDrafts + `) RETURNING id, storage_path`
	sentClient   = `EXISTS (SELECT 1 FROM documents k WHERE k.client_id = c.id AND k.status <> 'draft')`
	successor    = `(SELECT m.user_id FROM space_members m WHERE m.space_id = %s AND m.user_id <> $1
		ORDER BY (m.role IN ('owner', 'admin')) DESC, m.joined_at, m.id LIMIT 1)`
)

var (
	transferDrafts = `UPDATE documents d SET owner_id = COALESCE(` + fmt.Sprintf(successor, "d.space_id") + `, 0),
		external_ref = CASE WHEN d.external_ref <> '' AND EXISTS (
			SELECT 1 FROM documents x WHERE x.external_ref = d.external_ref
			AND x.owner_id = COALESCE(` + fmt.Sprintf(successor, "d.space_id") + `, 0)) THEN '' ELSE d.external_ref END
		WHERE d.owner_id = $1 AND d.space_id IS NOT NULL AND d.status = 'draft'`
	transferSpaces = `UPDATE spaces s SET owner_id = COALESCE(` + fmt.Sprintf(successor, "s.id") + `, 0) WHERE s.owner_id = $1`
)

var draftRows = []string{
	`DELETE FROM fields WHERE document_id IN (` + erasedDrafts + `)`,
	`DELETE FROM signers WHERE document_id IN (` + erasedDrafts + `)`,
}

var accountRows = []string{
	`DELETE FROM spaces s WHERE s.id IN (` + soleOwnedSpaces + `)
		AND NOT EXISTS (SELECT 1 FROM documents d WHERE d.space_id = s.id)`,
	`UPDATE documents SET client_id = NULL WHERE status = 'draft'
		AND client_id IN (SELECT c.id FROM clients c WHERE c.owner_id = $1 AND NOT ` + sentClient + `)`,
	`DELETE FROM clients c WHERE c.owner_id = $1 AND NOT ` + sentClient,
	`DELETE FROM webhooks WHERE owner_id = $1`,
	`DELETE FROM smtp_configs WHERE owner_id = $1`,
	transferDrafts,
	transferSpaces,
}

type erasedFiles struct {
	userID      int64
	avatarPath  string
	storagePath []string
	documentIDs []int64
}

// ErasePerson erases the account behind a Registre identity, following
// docs/erasure.md. Documents that were sent for signature, with their signers,
// fields, files and audit trail, are legal records and are never touched: the
// account is anonymised around them. Drafts that were never sent, the SMTP
// settings, webhooks, contacts no sent document uses, memberships, sessions,
// API tokens, pending MCP actions and the account row go. Drafts in a space others use pass to the
// space's longest-standing admin, else its longest-standing member, as the
// space itself does. Before anything is deleted, every completed or declined
// document's audit trail is written to disk while it can still print the
// owner's name and email. A subject that is already gone is a no-op, so a
// replayed notice changes nothing.
func (d *Deletions) ErasePerson(ctx context.Context, provider, subject string) error {
	if err := d.freezeFinalDocuments(ctx, provider, subject); err != nil {
		return err
	}
	var erased erasedFiles
	store := *d.users
	store.EraseData = func(ctx context.Context, tx *sql.Tx, userID int64) error {
		files, err := eraseAccountData(ctx, tx, userID)
		erased = files
		return err
	}
	if err := store.EraseAccount(ctx, provider, subject); err != nil {
		return err
	}
	d.removeFiles(erased)
	return nil
}

func (d *Deletions) freezeFinalDocuments(ctx context.Context, provider, subject string) error {
	if d.freeze == nil {
		return nil
	}
	trimmed := strings.TrimSuffix(provider, "/")
	rows, err := d.users.DB.QueryContext(ctx, finalDocuments, trimmed, trimmed+"/", subject)
	if err != nil {
		return fmt.Errorf("list the documents to freeze: %w", err)
	}
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, id := range ids {
		if err := d.freeze(ctx, id); err != nil {
			return fmt.Errorf("freeze the audit trail of document %d: %w", id, err)
		}
	}
	return nil
}

func eraseAccountData(ctx context.Context, tx *sql.Tx, userID int64) (erasedFiles, error) {
	files := erasedFiles{userID: userID}
	err := tx.QueryRowContext(ctx, `SELECT coalesce(avatar_upload_path, '') FROM users WHERE id = $1`, userID).Scan(&files.avatarPath)
	if err != nil {
		return files, fmt.Errorf("read the avatar of account %d: %w", userID, err)
	}
	if err := execAll(ctx, tx, draftRows, userID); err != nil {
		return files, err
	}
	if err := files.collectDrafts(ctx, tx); err != nil {
		return files, err
	}
	if err := mcpkit.ErasePending(ctx, tx, strconv.FormatInt(userID, 10)); err != nil {
		return files, err
	}
	return files, execAll(ctx, tx, accountRows, userID)
}

func execAll(ctx context.Context, tx *sql.Tx, statements []string, userID int64) error {
	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement, userID); err != nil {
			return fmt.Errorf("erase account %d: %w", userID, err)
		}
	}
	return nil
}

func (f *erasedFiles) collectDrafts(ctx context.Context, tx *sql.Tx) error {
	rows, err := tx.QueryContext(ctx, deleteDrafts, f.userID)
	if err != nil {
		return fmt.Errorf("erase the drafts of account %d: %w", f.userID, err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var path string
		if err := rows.Scan(&id, &path); err != nil {
			return err
		}
		f.documentIDs = append(f.documentIDs, id)
		f.storagePath = append(f.storagePath, path)
	}
	return rows.Err()
}

func (d *Deletions) removeFiles(erased erasedFiles) {
	var paths []string
	for i, id := range erased.documentIDs {
		if stored := erased.storagePath[i]; stored != "" {
			original := filepath.Join(d.uploadDir, filepath.Clean("/"+stored))
			paths = append(paths, original, strings.TrimSuffix(original, ".pdf")+"_signed.pdf")
		}
		paths = append(paths,
			filepath.Join(d.uploadDir, "audit", fmt.Sprintf("audit_%d.pdf", id)),
			filepath.Join(d.uploadDir, "certificates", fmt.Sprintf("cert_%d.pdf", id)))
	}
	if erased.avatarPath != "" {
		paths = append(paths, filepath.Join(d.uploadDir, filepath.Clean("/"+erased.avatarPath)))
	}
	for _, path := range paths {
		removeFile(path)
	}
	if erased.userID != 0 {
		_ = os.Remove(filepath.Join(d.uploadDir, fmt.Sprint(erased.userID)))
	}
}

func removeFile(path string) {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		slog.Warn("failed to remove an erased person's file", slog.Any("error", err))
	}
}

package secrets

import (
	"context"
	"log/slog"

	portepg "github.com/FacileStudio/porte/pg"
	"gorm.io/gorm"
)

const convertBatch = 500

const signersToConvert = `token IS NOT NULL AND token <> '' AND (token_hash IS NULL OR token_hash = '' OR (? AND token NOT LIKE 'pv1:%'))`

// Counts is how many rows of each kind one conversion pass changed, and how
// many signer tokens it left alone because no configured key opens them.
type Counts struct {
	Signers    int
	SMTP       int
	Webhooks   int
	Unreadable int
}

type pending struct {
	ID    int64
	Value string
}

type converter struct {
	ctx    context.Context
	db     *gorm.DB
	sealer *portepg.TokenSealer
	counts Counts
}

// Convert brings rows written before this release up to date: it fills
// token_hash for every signer that lacks it and, when a key is configured,
// seals signers.token, smtp_configs.password and webhooks.secret. Each row is
// rewritten by a conditional update on the value it read, so a second replica
// running the same pass changes nothing twice and a re-run changes nothing at
// all. It logs counts, never values.
func Convert(ctx context.Context, db *gorm.DB, sealer *portepg.TokenSealer, logger *slog.Logger) (Counts, error) {
	c := &converter{ctx: ctx, db: db, sealer: sealer}
	err := c.run()
	logger.Info("secrets at rest converted",
		slog.Int("signers", c.counts.Signers),
		slog.Int("smtp_configs", c.counts.SMTP),
		slog.Int("webhooks", c.counts.Webhooks),
		slog.Int("unreadable", c.counts.Unreadable),
		slog.Bool("sealing", sealer != nil),
	)
	return c.counts, err
}

func (c *converter) run() error {
	if err := c.batches("signers", "token", signersToConvert, c.signer, c.sealer != nil); err != nil {
		return err
	}
	if c.sealer == nil {
		return nil
	}
	if err := c.sealColumn("smtp_configs", "password", &c.counts.SMTP); err != nil {
		return err
	}
	return c.sealColumn("webhooks", "secret", &c.counts.Webhooks)
}

func (c *converter) batches(table, column, where string, handle func(pending) error, args ...any) error {
	var last int64
	for {
		var rows []pending
		err := c.db.WithContext(c.ctx).Table(table).Select("id, "+column+" AS value").
			Where("id > ?", last).Where(where, args...).Order("id").Limit(convertBatch).Scan(&rows).Error
		if err != nil || len(rows) == 0 {
			return err
		}
		for _, row := range rows {
			last = row.ID
			if err := handle(row); err != nil {
				return err
			}
		}
	}
}

func (c *converter) open(stored string) (string, bool) {
	plain, err := c.sealer.Open(stored)
	return plain, err == nil
}

func (c *converter) signer(row pending) error {
	plain, readable := c.open(row.Value)
	if !readable {
		c.counts.Unreadable++
		return nil
	}
	stored, hash, err := SealToken(c.sealer, plain)
	if err != nil {
		return err
	}
	if Sealed(row.Value) {
		stored = row.Value
	}
	res := c.db.WithContext(c.ctx).Table("signers").
		Where("id = ? AND token = ?", row.ID, row.Value).
		Updates(map[string]any{"token": stored, "token_hash": hash})
	c.counts.Signers += int(res.RowsAffected)
	return res.Error
}

func (c *converter) sealColumn(table, column string, changed *int) error {
	where := column + " IS NOT NULL AND " + column + " <> '' AND " + column + " NOT LIKE 'pv1:%'"
	return c.batches(table, column, where, func(row pending) error {
		sealed, err := c.sealer.Seal(row.Value)
		if err != nil {
			return err
		}
		res := c.db.WithContext(c.ctx).Table(table).Where("id = ? AND "+column+" = ?", row.ID, row.Value).Update(column, sealed)
		*changed += int(res.RowsAffected)
		return res.Error
	})
}

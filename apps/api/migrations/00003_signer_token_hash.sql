-- +goose Up
-- A signing link is looked up by the SHA-256 hex of its token, so signers.token can hold the sealed token. Rows written before this column are filled by the boot conversion pass, not here: sealing needs PORTE_TOKEN_KEY, which SQL cannot read.

ALTER TABLE signers ADD COLUMN token_hash text;
CREATE INDEX idx_signers_token_hash ON signers USING btree (token_hash);

-- +goose Down
DROP INDEX idx_signers_token_hash;
ALTER TABLE signers DROP COLUMN token_hash;

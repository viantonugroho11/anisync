-- Sequence untuk fencing token
CREATE SEQUENCE IF NOT EXISTS anisync_fencing_token_seq;

-- Tabel lock
CREATE TABLE IF NOT EXISTS anisync_locks (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  token BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_anisync_locks_expires_at ON anisync_locks (expires_at);



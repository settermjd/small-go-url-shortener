-- migrate:up
CREATE index IF NOT EXISTS idx_dates ON urls (created, updated);

-- migrate:down
DROP INDEX IF EXISTS idx_dates;
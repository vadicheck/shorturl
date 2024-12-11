CREATE TABLE IF NOT EXISTS urls
(
    id        INTEGER PRIMARY KEY,
    code      TEXT NOT NULL UNIQUE,
    url       TEXT NOT NULL UNIQUE
);
CREATE INDEX IF NOT EXISTS idx_code ON urls (code);

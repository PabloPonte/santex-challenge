CREATE TABLE features (
    name TEXT PRIMARY KEY,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL CHECK (status IN ('open', 'closed', 'whitelisted')),
    status_date TIMESTAMPTZ NOT NULL,
    whitelist JSONB NOT NULL DEFAULT '[]'::jsonb
);

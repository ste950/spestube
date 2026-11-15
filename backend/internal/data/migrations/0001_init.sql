-- Le chiavi di channels e videos sono gli identificativi di YouTube: sono già unici e
-- stabili, e un intero di comodo accanto sarebbe solo una seconda chiave da tenere allineata.

CREATE TABLE channels (
    id               TEXT    PRIMARY KEY,
    title            TEXT    NOT NULL,
    handle           TEXT    NOT NULL DEFAULT '',
    feed_mode        TEXT    NOT NULL DEFAULT 'uulf'
                             CHECK (feed_mode IN ('uulf', 'channel')),
    added_at         INTEGER NOT NULL,              -- secondi Unix, UTC
    last_fetched_at  INTEGER,                       -- NULL finché il feed non è mai stato letto
    last_fetch_error TEXT    NOT NULL DEFAULT ''
);

CREATE TABLE videos (
    id               TEXT    PRIMARY KEY,
    channel_id       TEXT    NOT NULL REFERENCES channels (id) ON DELETE CASCADE,
    title            TEXT    NOT NULL,
    published_at     INTEGER NOT NULL,
    fetched_at       INTEGER NOT NULL,
    source           TEXT    NOT NULL CHECK (source IN ('uulf', 'channel')),
    duration_seconds INTEGER,                       -- NULL senza chiave API
    is_live          INTEGER NOT NULL DEFAULT 0 CHECK (is_live IN (0, 1)),
    enriched_at      INTEGER,
    status           TEXT    NOT NULL
                             CHECK (status IN ('new', 'released', 'queued', 'watched', 'discarded', 'archived')),
    discard_reason   TEXT    NOT NULL DEFAULT ''
                             CHECK (discard_reason IN ('', 'user', 'baseline', 'short', 'live', 'duration', 'keyword')),
    released_at      INTEGER,
    queue_reason     TEXT    NOT NULL DEFAULT '',
    queued_at        INTEGER,
    watched_at       INTEGER,
    archived_at      INTEGER,
    -- In coda si entra solo con un motivo: lo garantisce il database, non solo il servizio.
    CHECK (status <> 'queued' OR (trim(queue_reason) <> '' AND queued_at IS NOT NULL))
);

CREATE INDEX videos_status_released ON videos (status, released_at);
CREATE INDEX videos_status_fetched ON videos (status, fetched_at);
CREATE INDEX videos_status_queued ON videos (status, queued_at);
CREATE INDEX videos_channel ON videos (channel_id);

CREATE TABLE notes (
    video_id   TEXT    PRIMARY KEY REFERENCES videos (id) ON DELETE CASCADE,
    body       TEXT    NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE TABLE settings (
    key        TEXT    PRIMARY KEY,
    value      TEXT    NOT NULL,
    updated_at INTEGER NOT NULL
);

-- Le raccolte prendono il posto di Da vedere: si salvano video in raccolte con un nome,
-- senza motivo obbligatorio e senza scadenza.

CREATE TABLE collections (
    id          INTEGER PRIMARY KEY,
    -- Due raccolte con lo stesso nome, maiuscole a parte, sarebbero impossibili da
    -- distinguere nella lista.
    name        TEXT    NOT NULL COLLATE NOCASE UNIQUE,
    description TEXT    NOT NULL DEFAULT '',
    position    INTEGER NOT NULL,
    created_at  INTEGER NOT NULL,
    updated_at  INTEGER NOT NULL
);

CREATE INDEX collections_position ON collections (position);

CREATE TABLE collection_items (
    collection_id INTEGER NOT NULL REFERENCES collections (id) ON DELETE CASCADE,
    video_id      TEXT    NOT NULL REFERENCES videos (id) ON DELETE CASCADE,
    added_at      INTEGER NOT NULL,
    PRIMARY KEY (collection_id, video_id)
);

CREATE INDEX collection_items_video ON collection_items (video_id);

-- I video che erano in Da vedere non si perdono: finiscono in una raccolta con lo stesso
-- nome, nell'ordine in cui erano entrati, e tornano a essere video archiviati.
INSERT INTO collections (name, description, position, created_at, updated_at)
SELECT 'Da vedere', '', 0, unixepoch(), unixepoch()
WHERE EXISTS (SELECT 1 FROM videos WHERE status = 'queued');

INSERT INTO collection_items (collection_id, video_id, added_at)
SELECT (SELECT id FROM collections WHERE name = 'Da vedere'), id, COALESCE(queued_at, unixepoch())
FROM videos
WHERE status = 'queued';

-- Lo stato 'queued' e le colonne queue_reason e queued_at restano nello schema perché un
-- vincolo CHECK le nomina, e in SQLite toglierle vorrebbe dire ricostruire la tabella dei
-- video. Da qui in avanti nessuno le scrive più.
UPDATE videos SET status = 'archived', archived_at = unixepoch() WHERE status = 'queued';

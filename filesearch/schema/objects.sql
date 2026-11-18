-- filesearch.volume
-- A folder which is indexed, by its absolute path, with the status of its
-- indexing. Its run is counted, from one, as each indexing of it starts,
-- which marks the objects it sees, so those it didn't are swept
CREATE TABLE IF NOT EXISTS volume (
    id        INTEGER PRIMARY KEY,
    path      TEXT NOT NULL UNIQUE CHECK (path <> ''),  -- absolute, with / separators
    recursive BOOLEAN NOT NULL DEFAULT TRUE,             -- its folders are indexed, as well as its files
    status    TEXT NOT NULL DEFAULT 'new'
              CHECK (status IN ('new', 'indexing', 'indexed', 'failed')),
    run       INTEGER NOT NULL DEFAULT 0,               -- the run, which is 0 until it's first indexed
    started   TEXT,                                     -- when the run started, in UTC, of RFC 3339
    finished  TEXT,                                     -- when it finished, or NULL while it's indexing
    error     TEXT                                      -- why it failed, or NULL
);

-- filesearch.object
-- A file, of a volume, by its path within it, with its metadata, which the
-- run which last saw it marks
CREATE TABLE IF NOT EXISTS object (
    id       INTEGER PRIMARY KEY,
    volume   INTEGER NOT NULL REFERENCES volume (id) ON DELETE CASCADE,
    path     TEXT NOT NULL CHECK (path <> ''),                    -- within the volume, with / separators
    size     INTEGER NOT NULL CHECK (size >= 0),                  -- in bytes
    modtime  TEXT NOT NULL,                                       -- when it was modified, in UTC, of RFC 3339
    type     TEXT,                                                -- its content type, such as image/jpeg
    kind     TEXT,                                                -- its kind, such as photo or book
    title    TEXT,                                                -- its title
    meta     TEXT NOT NULL DEFAULT '{}' CHECK (json_valid(meta)),  -- its metadata, of "ns:key", such as "exif:Make"
    artwork  BLOB,                                                -- an image, such as a thumbnail or a cover
    indexed  TEXT NOT NULL,                                       -- when it was indexed, in UTC, of RFC 3339
    run      INTEGER NOT NULL,                                    -- the run of its volume which last saw it
    UNIQUE (volume, path)
);

-- filesearch.object_run
CREATE INDEX IF NOT EXISTS object_run ON object (volume, run);

-- filesearch.object_text
-- The text of each object, by its id, which is searched, and not kept: its
-- title, the values of its metadata, and its content, as Markdown
CREATE VIRTUAL TABLE IF NOT EXISTS object_text USING fts5(
    title, meta, text,
    content = '', contentless_delete = 1
);

-- filesearch.object_delete
-- An object's text is deleted with it, as it is with its volume
CREATE TRIGGER IF NOT EXISTS object_delete AFTER DELETE ON object BEGIN
    DELETE FROM object_text WHERE rowid = old.id;
END;

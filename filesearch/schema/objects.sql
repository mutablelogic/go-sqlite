-- filesearch.object
-- A file, of a root folder, by its absolute path, with its metadata, which
-- the run which last saw it marks, so the files which a complete run of
-- the root didn't see are swept
CREATE TABLE IF NOT EXISTS object (
    id       INTEGER PRIMARY KEY,
    path     TEXT NOT NULL UNIQUE CHECK (path <> ''),            -- absolute, with / separators
    size     INTEGER NOT NULL CHECK (size >= 0),                  -- in bytes
    modtime  TEXT NOT NULL,                                       -- when it was modified, in UTC, of RFC 3339
    type     TEXT,                                                -- its content type, such as image/jpeg
    kind     TEXT,                                                -- its kind, such as photo or book
    title    TEXT,                                                -- its title
    meta     TEXT NOT NULL DEFAULT '{}' CHECK (json_valid(meta)),  -- its metadata, of "ns:key", such as "exif:Make"
    artwork  BLOB,                                                -- an image, such as a thumbnail or a cover
    indexed  TEXT NOT NULL,                                       -- when it was indexed, in UTC, of RFC 3339
    run      INTEGER NOT NULL                                     -- the run which last saw it
);

-- filesearch.object_run
CREATE INDEX IF NOT EXISTS object_run ON object (run);

-- filesearch.object_text
-- The text of each object, by its id, which is searched, and not kept: its
-- title, the values of its metadata, and its content, as Markdown
CREATE VIRTUAL TABLE IF NOT EXISTS object_text USING fts5(
    title, meta, text,
    content = '', contentless_delete = 1
);

-- filesearch.object_delete
-- An object's text is deleted with it
CREATE TRIGGER IF NOT EXISTS object_delete AFTER DELETE ON object BEGIN
    DELETE FROM object_text WHERE rowid = old.id;
END;

-- filesearch.run
-- A run of the indexer, of a root folder, which is complete once it's
-- finished
CREATE TABLE IF NOT EXISTS run (
    id        INTEGER PRIMARY KEY,
    root      TEXT NOT NULL CHECK (root <> ''),  -- absolute, with / separators
    started   TEXT NOT NULL,                      -- in UTC, of RFC 3339
    finished  TEXT                                -- in UTC, of RFC 3339, or NULL while it runs
);

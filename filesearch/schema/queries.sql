-- filesearch.run_start
INSERT INTO run (root, started) VALUES (@root, @started)
RETURNING id, root, started, finished;

-- filesearch.run_finish
UPDATE run SET finished = @finished WHERE id = @id
RETURNING id, root, started, finished;

-- filesearch.run_get
SELECT id, root, started, finished FROM run WHERE id = @id;

-- filesearch.object_mark
-- An object which is as it was, of its size and when it was modified, is
-- marked as seen by the run, without indexing it again
UPDATE object SET run = @run
WHERE path = @path AND size = @size AND modtime = @modtime
RETURNING id, path, size, modtime, type, kind, title, meta, indexed, run;

-- filesearch.object_upsert
-- An object, and its text, which replaces it, of its path
INSERT INTO object (path, size, modtime, type, kind, title, meta, artwork, indexed, run)
VALUES (@path, @size, @modtime, @type, @kind, @title, @meta, @artwork, @indexed, @run)
ON CONFLICT (path) DO UPDATE SET
    size = excluded.size, modtime = excluded.modtime, type = excluded.type,
    kind = excluded.kind, title = excluded.title, meta = excluded.meta,
    artwork = excluded.artwork, indexed = excluded.indexed, run = excluded.run
RETURNING id, path, size, modtime, type, kind, title, meta, indexed, run;
DELETE FROM object_text WHERE rowid = (SELECT id FROM object WHERE path = @path);
INSERT INTO object_text (rowid, title, meta, text)
VALUES ((SELECT id FROM object WHERE path = @path), @title, @meta_text, @text);

-- filesearch.object_get
SELECT id, path, size, modtime, type, kind, title, meta, indexed, run
FROM object WHERE id = @id;

-- filesearch.object_delete
DELETE FROM object WHERE id = @id
RETURNING id, path, size, modtime, type, kind, title, meta, indexed, run;

-- filesearch.object_sweep
-- The objects of a root, which a complete run of it didn't see
DELETE FROM object
WHERE run < @run AND (path = @root OR substr(path, 1, length(@prefix)) = @prefix)
RETURNING id, path, size, modtime, type, kind, title, meta, indexed, run;

-- filesearch.object_list
SELECT id, path, size, modtime, type, kind, title, meta, indexed, run
FROM object ${where} ORDER BY path

-- filesearch.object_search
-- Objects whose text matches, the best first, of their titles, then their
-- metadata, then their content
SELECT object.id, path, size, modtime, type, kind, object.title, object.meta, indexed, run
FROM object_text JOIN object ON object.id = object_text.rowid
WHERE object_text MATCH @match ${and}
ORDER BY bm25(object_text, 10.0, 2.0, 1.0)

-- filesearch.artwork_get
SELECT artwork FROM object WHERE id = @id AND artwork IS NOT NULL;

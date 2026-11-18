-- filesearch.volume_insert
-- A volume, of its path, or the one which is of it already, which is
-- recursive, or not, as it's indexed next
INSERT INTO volume (path, recursive) VALUES (@path, @recursive)
ON CONFLICT (path) DO UPDATE SET recursive = excluded.recursive
RETURNING id, path, recursive, status, run, started, finished, error;

-- filesearch.volume_get
SELECT id, path, recursive, status, run, started, finished, error FROM volume WHERE id = @id;

-- filesearch.volume_list
SELECT id, path, recursive, status, run, started, finished, error FROM volume ORDER BY path

-- filesearch.volume_start
-- Its next run starts, which is indexing it
UPDATE volume SET status = 'indexing', run = run + 1, started = @started, finished = NULL, error = NULL
WHERE id = @id
RETURNING id, path, recursive, status, run, started, finished, error;

-- filesearch.volume_finish
-- Its run finishes, which indexed it, or failed, with its error
UPDATE volume SET status = @status, finished = @finished, error = @error
WHERE id = @id AND status = 'indexing'
RETURNING id, path, recursive, status, run, started, finished, error;

-- filesearch.volume_delete
-- A volume, with its objects, and their text
DELETE FROM volume WHERE id = @id
RETURNING id, path, recursive, status, run, started, finished, error;

-- filesearch.object_mark
-- An object which is as it was, of its size and when it was modified, is
-- marked as seen by the run, without indexing it again
UPDATE object SET run = @run
WHERE volume = @volume AND path = @path AND size = @size AND modtime = @modtime
RETURNING id, volume, path, size, modtime, type, kind, title, meta, indexed, run;

-- filesearch.object_upsert
-- An object, and its text, which replaces it, of its volume and path
INSERT INTO object (volume, path, size, modtime, type, kind, title, meta, artwork, indexed, run)
VALUES (@volume, @path, @size, @modtime, @type, @kind, @title, @meta, @artwork, @indexed, @run)
ON CONFLICT (volume, path) DO UPDATE SET
    size = excluded.size, modtime = excluded.modtime, type = excluded.type,
    kind = excluded.kind, title = excluded.title, meta = excluded.meta,
    artwork = excluded.artwork, indexed = excluded.indexed, run = excluded.run
RETURNING id, volume, path, size, modtime, type, kind, title, meta, indexed, run;
DELETE FROM object_text WHERE rowid = (SELECT id FROM object WHERE volume = @volume AND path = @path);
INSERT INTO object_text (rowid, title, meta, text)
VALUES ((SELECT id FROM object WHERE volume = @volume AND path = @path), @title, @meta_text, @text);

-- filesearch.object_get
SELECT id, volume, path, size, modtime, type, kind, title, meta, indexed, run
FROM object WHERE id = @id;

-- filesearch.object_delete
DELETE FROM object WHERE id = @id
RETURNING id, volume, path, size, modtime, type, kind, title, meta, indexed, run;

-- filesearch.object_sweep
-- The objects of a volume, which its run didn't see
DELETE FROM object WHERE volume = @volume AND run < @run
RETURNING id, volume, path, size, modtime, type, kind, title, meta, indexed, run;

-- filesearch.object_list
SELECT id, volume, path, size, modtime, type, kind, title, meta, indexed, run
FROM object ${where} ORDER BY volume, path

-- filesearch.object_search
-- Objects whose text matches, the best first, of their titles, then their
-- metadata, then their content
SELECT object.id, volume, path, size, modtime, type, kind, object.title, object.meta, indexed, run
FROM object_text JOIN object ON object.id = object_text.rowid
WHERE object_text MATCH @match ${and}
ORDER BY bm25(object_text, 10.0, 2.0, 1.0)

-- filesearch.artwork_get
SELECT artwork FROM object WHERE id = @id AND artwork IS NOT NULL;

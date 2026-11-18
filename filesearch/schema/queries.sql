-- filesearch.volume_insert
-- A volume, of its path, or the one which is of it already
INSERT INTO volume (path) VALUES (@path)
ON CONFLICT (path) DO UPDATE SET path = excluded.path
RETURNING id, path, status, run, started, finished, error;

-- filesearch.volume_get
SELECT id, path, status, run, started, finished, error FROM volume WHERE id = @id;

-- filesearch.volume_list
SELECT id, path, status, run, started, finished, error FROM volume ORDER BY path

-- filesearch.volume_start
-- Its next run starts, which is indexing it
UPDATE volume SET status = 'indexing', run = run + 1, started = @started, finished = NULL, error = NULL
WHERE id = @id
RETURNING id, path, status, run, started, finished, error;

-- filesearch.volume_finish
-- Its run finishes, which indexed it, or failed, with its error
UPDATE volume SET status = @status, finished = @finished, error = @error
WHERE id = @id AND status = 'indexing'
RETURNING id, path, status, run, started, finished, error;

-- filesearch.volume_delete
-- A volume, with its objects, and their text
DELETE FROM volume WHERE id = @id
RETURNING id, path, status, run, started, finished, error;

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
INSERT INTO object_text (rowid, title, path, meta, text)
VALUES ((SELECT id FROM object WHERE volume = @volume AND path = @path), @title, @path_text, @meta_text, @text);

-- filesearch.object_get
SELECT id, volume, path, size, modtime, type, kind, title, meta, indexed, run
FROM object WHERE id = @id;

-- filesearch.object_delete
DELETE FROM object WHERE id = @id
RETURNING id, volume, path, size, modtime, type, kind, title, meta, indexed, run;

-- filesearch.object_sweep
-- The objects of a volume, which its run didn't see, once its run indexed it,
-- and only of its latest run, so a run which is indexing, or failed, or an
-- older run, sweeps nothing
DELETE FROM object
WHERE volume = @volume AND run < @run
AND EXISTS (SELECT 1 FROM volume WHERE id = @volume AND run = @run AND status = 'indexed')
RETURNING id, volume, path, size, modtime, type, kind, title, meta, indexed, run, NULL AS rank;

-- filesearch.object_list
SELECT id, volume, path, size, modtime, type, kind, title, meta, indexed, run, NULL AS rank
FROM object ${where} ORDER BY ${order}

-- filesearch.object_search
-- Objects whose text matches, of their rank, between 0 and 1, which is higher
-- the better they match, of their titles, then their paths, then their
-- metadata, then their content, of a minimum, or not. The rank is bm25's
-- score, which is higher the better, and isn't bounded, as score / (1 + score)
SELECT * FROM (
    SELECT object.id, volume, object.path, size, modtime, type, kind, object.title, object.meta, indexed, run,
           score / (1.0 + score) AS rank
    FROM (
        SELECT rowid, max(-bm25(object_text, 10.0, 5.0, 2.0, 1.0), 0.0) AS score
        FROM object_text WHERE object_text MATCH @match
    ) AS matched JOIN object ON object.id = matched.rowid
    ${where}
) ${rank} ORDER BY ${order}

-- filesearch.artwork_get
SELECT artwork FROM object WHERE id = @id AND artwork IS NOT NULL;

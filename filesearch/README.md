# filesearch

Indexes folders, of their files' metadata, content and artwork, with [go-metadata](https://github.com/mutablelogic/go-metadata), and searches them, in a SQLite database.

- **Volumes:** a folder which is indexed. Indexing it again marks the files which are as they were, reads those which are new or changed, and sweeps what's gone.
- **Search:** of the words in files' titles, paths, metadata and content, as prefixes, with SQLite's full-text search. Each file has a rank, between 0 and 1, which is higher the better it matches, of its title, then its path, then its metadata, then its content, and those under a cutoff, of 0.15 by default, aren't found.
- **Its own module**, so go-sqlite's dependencies stay small.

## The command

```bash
make filesearch                  # build/filesearch: documents, email, text and images
make filesearch MEDIA=1          # and audio, video and photos' EXIF, HEIF and RAW, with cgo
```

`MEDIA=1` needs cgo, and the C libraries which go-metadata reads them with: FFmpeg 8, libexif, libheif and LibRaw (see go-metadata's README).

```bash
filesearch index ~/Documents     # Index a folder, of its folders too, with --no-recursive not
filesearch search tax return     # Search, of all of the words, the best first, of a rank of 0.15 or more
filesearch search --cutoff 0.5 tax  # Of a rank of 0.5, or more, or 0, for any
filesearch search --order date tax  # The newest first, rather than the best
filesearch search                # List the files, the newest first
filesearch search --order size   # The largest first, or by path, or title
filesearch search --type video   # Of a kind, or the start, or end, of a content type, such as video, or pdf
filesearch search --type video/mp4 rain  # Of a content type
filesearch volumes               # The folders, with the status of their indexing
```

The database is `filesearch.db`, in the user's cache folder, or `--db`.

Hidden files and folders, lock files, backups and autosaves, temporary files and partial downloads, the files of Windows and of file systems, and dependencies' folders aren't indexed, unless `--no-default-exclude`. `--exclude` excludes others, by a regular expression of their names, such as `--exclude '\.bak$'`.

## The package

```go
manager, err := filesearch.New(ctx, path, filesearch.WithTracer(tracer))
volume, err := manager.IndexVolume(ctx, schema.IndexVolumeRequest{Path: folder, Recursive: true}, progress)
list, err := manager.ListObjects(ctx, schema.ObjectListRequest{Query: types.Ptr("tax return"), Cutoff: types.Ptr(0.5)})
```

- **Excluded:** `WithExclude(patterns...)` excludes files and folders by regular expressions of their names, as well as `DefaultExclude`, unless `WithoutDefaultExclude()`. What was indexed, and then is excluded, is swept.
- **`ProgressFn`** is called as each batch of files is committed, with a file's error, when it can't be read, and once it's finished.
- **`Progress(volume)`** is how far a run has indexed a volume, or how it did.
- **Handlers:** the program which uses it imports go-metadata's handlers, which read files, of their types, such as `_ "github.com/mutablelogic/go-metadata/text"`. `cmd/filesearch` imports those which are pure Go, and, with the `media` tag, the others.

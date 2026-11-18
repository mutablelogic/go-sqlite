package filesearch

import (
	"regexp"

	// Packages
	tmdb "github.com/mutablelogic/go-metadata/pkg/tmdb"
	sqlite "github.com/mutablelogic/go-sqlite"
	trace "go.opentelemetry.io/otel/trace"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Opt is an option of a manager, when it's created
type Opt func(*opt) error

// opt is the options of a manager
type opt struct {
	tracer         trace.Tracer
	readers        int
	exclude        []*regexp.Regexp // Of the names of files and folders, which aren't indexed
	defaultExclude bool             // DefaultExclude, as well
	tmdb           *tmdb.Client     // Of the movies and episodes on TMDB which videos are, or nil
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

// DefaultExclude are the names of files and folders which aren't indexed,
// unless WithoutDefaultExclude is an option: hidden files and folders, such
// as .git and .DS_Store, Office's lock files, editors' backups and
// autosaves, temporary files and partial downloads, the files and folders
// of Windows and of file systems, and the folders of dependencies and
// builds
var DefaultExclude = []string{
	`^\.`,
	`^~\$`,
	`~$`,
	`^#.*#$`,
	`(?i)\.(tmp|temp|swp|swo|part|crdownload|download)$`,
	`(?i)^(thumbs\.db|desktop\.ini|ehthumbs\.db)$`,
	`(?i)^(\$recycle\.bin|system volume information|lost\+found)$`,
	`^(node_modules|__pycache__)$`,
	`\.pyc$`,
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (o *opt) apply(opts []Opt) error {
	o.readers = sqlite.DefaultReaders
	o.defaultExclude = true
	for _, fn := range opts {
		if err := fn(o); err != nil {
			return err
		}
	}
	if o.defaultExclude {
		for _, pattern := range DefaultExclude {
			o.exclude = append(o.exclude, regexp.MustCompile(pattern))
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// WithTracer sets the OpenTelemetry tracer, of the spans of the manager,
// and of its database's queries
func WithTracer(tracer trace.Tracer) Opt {
	return func(o *opt) error {
		o.tracer = tracer
		return nil
	}
}

// WithMaxReaders sets the most connections which read the database at
// once, such as to search it, while it's indexed
func WithMaxReaders(n int) Opt {
	return func(o *opt) error {
		if n < 1 {
			return sqlite.ErrBadParameter.Withf("readers: %d", n)
		}
		o.readers = n
		return nil
	}
}

// WithExclude excludes the files, and folders, whose names match the
// regular expressions, such as `\.bak$`, as well as DefaultExclude. An
// excluded folder's files aren't indexed, and what's excluded which was
// indexed is swept, when its volume is indexed again.
func WithExclude(patterns ...string) Opt {
	return func(o *opt) error {
		for _, pattern := range patterns {
			re, err := regexp.Compile(pattern)
			if err != nil {
				return sqlite.ErrBadParameter.Withf("exclude %q: %v", pattern, err)
			}
			o.exclude = append(o.exclude, re)
		}
		return nil
	}
}

// WithTMDB looks up the movies and episodes on TMDB which videos are, of
// their names, with the client, so their titles, dates, descriptions and
// posters are TMDB's. A video is looked up when it's indexed, which is when
// it's new, or changed.
func WithTMDB(client *tmdb.Client) Opt {
	return func(o *opt) error {
		o.tmdb = client
		return nil
	}
}

// WithoutDefaultExclude doesn't exclude DefaultExclude, so hidden files,
// and the others, are indexed, unless WithExclude excludes them
func WithoutDefaultExclude() Opt {
	return func(o *opt) error {
		o.defaultExclude = false
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// excluded returns true when a file's, or folder's, name is excluded
func (o *opt) excluded(name string) bool {
	for _, re := range o.exclude {
		if re.MatchString(name) {
			return true
		}
	}
	return false
}

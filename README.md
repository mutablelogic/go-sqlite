# sqlite

Higher-level interface to SQLite. See the `sq_import` command in
order to understand how to use this component. In order to compile
the sq_import tool,

```bash
bash% cd sqlite
bash% make sq_import
```

The command line arguments are:

```

sq_import <flags> <csv_file>...

Flags:
  -noheader
    	Do not use the first row as column names
  -notnull
    	Dont use NULL values for empty values
  -skipcomments
    	Skip comment lines (default true) which start with # or //
  -sqlite.dsn string
    	Database source (default ":memory:")
  -verbose
    	Verbose logging
  -version
    	Print version information and exit
  -debug
    	Set debugging mode
```


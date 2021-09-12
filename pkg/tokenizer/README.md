# tokenizer package

This package parses SQL statements to do two specific things:

  * Returns tokens for SQL statements, where those tokens identify keyword, type, name and value
    text within the SQL statement;
  * Can determine if a statement "is complete - that is, has a trailing semicolon on a statement.

This package is part of a wider project, `github.com/djthorpe/go-sqlite`.
Please see the [module documentation](https://github.com/djthorpe/go-sqlite/blob/master/README.md)
for more information.

## Using the tokenizer

Here's an example of using the tokenizer:


```go
import (
	"github.com/djthorpe/go-sqlite/pkg/tokenizer"
)

func Tokenize(q string) ([]interface{},error) {
    tokenizer := NewTokenizer(test)
    tokens := []interface{}{}
    for {
        token, err := tokenizer.Next()
        if token == nil {
            return tokens, nil
        }
        if err != nil {
            return nil, err
        }
        tokens = append(tokens, token)
    }
}
```

Tokens returned can be one of the following types:

  * `KeywordToken`: a keyword, such as `SELECT`, `FROM`, `WHERE`, etc.
  * `TypeToken`: a type such as `INTEGER`, `TEXT`, etc
  * `NameToken`: a table or column name
  * `Value Token`: a numeric, boolean or text value
  * `WhitespaceToken`: Spaces, tabs and newlines
  * `PuncuationToken`: anything not included above

## Establishing if a statement is complete

Call the `func IsComplete(string) bool` method to determine if a statement is complete.
As per the [sqlite documentation](https://www.sqlite.org/c3ref/complete.html) "useful during 
command-line input to determine if the currently entered text seems to form a complete SQL 
statement or if additional input is needed before sending the text into SQLite for parsing".

However, "...do not parse the SQL statements thus will not detect syntactically incorrect SQL."



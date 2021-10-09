package lang

import (
	"fmt"
	"reflect"
	"strings"
)

// sliceJoin elements of a slice with a separator. If fn is nil then each element is
// evaluated with fmt.Sprint function
func sliceJoin(slice interface{}, sep string, fn func(elem interface{}) string) string {
	rv := reflect.ValueOf(slice)
	if rv.Kind() != reflect.Slice {
		panic("sliceJoin: not a slice")
	}
	if rv.Len() == 0 {
		return ""
	}
	result := make([]string, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		if fn == nil {
			result[i] = fmt.Sprint(rv.Index(i).Interface())
		} else {
			result[i] = fn(rv.Index(i).Interface())
		}
	}
	return strings.Join(result, sep)
}

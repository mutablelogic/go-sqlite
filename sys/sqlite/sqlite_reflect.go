/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package sqlite

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	sq "github.com/djthorpe/sqlite"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type tagflag int

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	TAG_NAME tagflag = iota
	TAG_NULLABLE
	TAG_DECLTYPE
)

////////////////////////////////////////////////////////////////////////////////
// REFLECT IMPLEMENTATION

func (this *sqlite) Reflect(v interface{}) ([]sq.Column, error) {
	this.log.Debug2("<sqlite.Reflect>{ %T }", v)

	// Dereference the pointer
	v_ := reflect.ValueOf(v)
	for v_.Kind() == reflect.Ptr {
		v_ = v_.Elem()
	}
	// If not a stuct then return
	if v_.Kind() != reflect.Struct {
		return nil, gopi.ErrBadParameter
	}
	// Enumerate struct fields
	columns := make([]sq.Column, 0, v_.Type().NumField())
	for i := 0; i < v_.Type().NumField(); i++ {
		if column, err := reflectField(v_, i, len(columns)); err != nil {
			return nil, err
		} else if column != nil {
			columns = append(columns, column)
		}
	}
	// Return columns
	return columns, nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func reflectField(v reflect.Value, i, pos int) (sq.Column, error) {
	if tags := reflectFieldTags(v, i); tags == nil {
		// Ignore if no tags returned, or private name
		return nil, nil
	} else if decltype, _ := tags[TAG_DECLTYPE]; decltype == "" {
		return nil, fmt.Errorf("Unsupported type for field %v", strconv.Quote(tags[TAG_NAME]))
	} else {
		_, nullable := tags[TAG_NULLABLE]
		this := &column{
			name:     tags[TAG_NAME],
			pos:      pos,
			nullable: nullable,
			decltype: decltype,
		}
		return this, nil
	}
}

func reflectFieldTags(v reflect.Value, i int) map[tagflag]string {
	f := v.Type().Field(i)
	tags := map[tagflag]string{
		TAG_NAME: f.Name,
	}
	if f.Anonymous == true || f.Name == "" {
		// No field name - return nil
		return nil
	} else if runes := []rune(f.Name); unicode.IsUpper(runes[0]) == false {
		// Ignore non-exported fields - return nil
		return nil
	} else if tag, ok := f.Tag.Lookup(DEFAULT_STRUCT_TAG); ok == false {
		// No tag
	} else if tag == "-" {
		// Ignore field - return nil
		return nil
	} else {
		name, options := tagParse(tag)
		if name != "" {
			tags[TAG_NAME] = name
		}
		// Check nullable
		if tagHasOption(options, "NULLABLE") {
			tags[TAG_NULLABLE] = "NULLABLE"
		}
		// Check types - choose first one from list of supported types
		for _, decltype := range sq.SupportedTypes() {
			if tagHasOption(options, decltype) {
				tags[TAG_DECLTYPE] = decltype
				break
			}
		}
	}

	// Set decltype if not already set
	if _, exists := tags[TAG_DECLTYPE]; exists == false {
		tags[TAG_DECLTYPE] = sq.SupportedTypeForType(v.Field(i).Interface())
	}

	// Return the tags
	return tags
}

func tagParse(tag string) (string, string) {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx], tag[idx+1:]
	}
	return tag, ""
}

func tagHasOption(tag, option string) bool {
	option = strings.ToUpper(option)
	for tag != "" {
		var next string
		i := strings.Index(tag, ",")
		if i >= 0 {
			tag, next = tag[:i], tag[i+1:]
		}
		if strings.ToUpper(tag) == option {
			return true
		}
		tag = next
	}
	return false
}

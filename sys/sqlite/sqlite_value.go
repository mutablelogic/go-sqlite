/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package sqlite

import (
	"fmt"
	"math"
	"reflect"
	"time"
)

////////////////////////////////////////////////////////////////////////////////
// VALUE IMPLEMENTATION

func (this *value) String() string {
	if this.v == nil {
		return ""
	}
	switch this.v.(type) {
	case string:
		return this.v.(string)
	default:
		return fmt.Sprint(this.v)
	}
}

func (this *value) Int() int64 {
	if this.v == nil {
		return 0
	}
	switch this.v.(type) {
	case bool:
		if this.v.(bool) {
			return -1
		} else {
			return 0
		}
	case int:
		return int64(this.v.(int))
	case int16:
		return int64(this.v.(int16))
	case int32:
		return int64(this.v.(int32))
	case int64:
		return int64(this.v.(int64))
	default:
		panic(fmt.Sprintf("Invalid type conversion: %v", reflect.TypeOf(this.v)))
	}
}

func (this *value) Uint() uint64 {
	if this.v == nil {
		return 0
	}
	switch this.v.(type) {
	case bool:
		if this.v.(bool) {
			return math.MaxUint64
		} else {
			return 0
		}
	case int:
		return uint64(this.v.(int))
	case int16:
		return uint64(this.v.(int16))
	case int32:
		return uint64(this.v.(int32))
	case int64:
		return uint64(this.v.(int64))
	default:
		panic(fmt.Sprintf("Invalid type conversion: %v", reflect.TypeOf(this.v)))
	}
}

func (this *value) IsNull() bool {
	if this.v == nil {
		return true
	} else {
		return false
	}
}

func (this *value) Bool() bool {
	if this.v == nil {
		return false
	}
	switch this.v.(type) {
	case bool:
		return this.v.(bool)
	case int:
		return this.v.(int) != 0
	case int16:
		return this.v.(int16) != 0
	case int32:
		return this.v.(int32) != 0
	case int64:
		return this.v.(int64) != 0
	default:
		panic(fmt.Sprintf("Invalid type conversion: %v", reflect.TypeOf(this.v)))
	}
}

func (this *value) Float() float64 {
	if this.v == nil {
		return 0
	}
	switch this.v.(type) {
	case float32:
		return float64(this.v.(float32))
	case float64:
		return float64(this.v.(float64))
	case int:
		return float64(this.v.(int))
	case int16:
		return float64(this.v.(int16))
	case int32:
		return float64(this.v.(int32))
	case int64:
		return float64(this.v.(int64))
	default:
		panic(fmt.Sprintf("Invalid type conversion: %v", reflect.TypeOf(this.v)))
	}
}

func (this *value) Timestamp() time.Time {
	if this.v == nil {
		return time.Time{}
	}
	switch this.v.(type) {
	case time.Time:
		return this.v.(time.Time)
	default:
		panic(fmt.Sprintf("Invalid type conversion: %v", reflect.TypeOf(this.v)))
	}
}

func (this *value) Bytes() []byte {
	if this.v == nil {
		return nil
	}
	switch this.v.(type) {
	case string:
		return []byte(this.v.(string))
	case []byte:
		return this.v.([]byte)
	default:
		panic(fmt.Sprintf("Invalid type conversion: %v", reflect.TypeOf(this.v)))
	}
}

func (this *value) DeclType() string {
	if this.c != nil {
		return this.c.decltype
	} else {
		return DEFAULT_COLUMN_TYPE
	}
}

package indexer

import (
	"fmt"
	"io/fs"
	"strings"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type IndexerEvent interface {
	Name() string
	Type() EventType
	FileInfo() fs.FileInfo
	Path() string
}

type event struct {
	EventType
	name string
	path string
	info fs.FileInfo
}

type EventType int

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	EVENT_TYPE_ADDED EventType = (1 << iota)
	EVENT_TYPE_REMOVED
	EVENT_TYPE_RENAMED
	EVENT_TYPE_CHANGED
	EVENT_TYPE_NONE EventType = 0
	EVENT_TYPE_MIN            = EVENT_TYPE_ADDED
	EVENT_TYPE_MAX            = EVENT_TYPE_CHANGED
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewEvent(t EventType, name, path string, info fs.FileInfo) *event {
	return &event{
		EventType: t,
		name:      name,
		path:      path,
		info:      info,
	}
}

///////////////////////////////////////////////////////////////////////////////
// EVENT IMPLEMENTATION

func (e *event) Name() string {
	return e.name
}

func (e *event) Value() interface{} {
	return e.path
}

func (e *event) Type() EventType {
	return e.EventType
}

func (e *event) FileInfo() fs.FileInfo {
	return e.info
}

func (e *event) Path() string {
	return e.path
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (e *event) String() string {
	str := "<indexer.event"
	if t := e.EventType; t != EVENT_TYPE_NONE {
		str += " type=" + e.EventType.String()
	}
	if name := e.Name(); name != "" {
		str += fmt.Sprintf(" name=%q", name)
	}
	if path := e.Value(); path != nil {
		str += fmt.Sprintf(" path=%q", path)
	}
	if e.info == nil {
		str += " info=<nil>"
	} else {
		str += " info={FileInfo}"
	}
	return str + ">"
}

func (f EventType) String() string {
	if f == EVENT_TYPE_NONE {
		return f.FlagString()
	}
	str := ""
	for v := EVENT_TYPE_MIN; v <= EVENT_TYPE_MAX; v <<= 1 {
		if f&v == v {
			str += v.FlagString() + "|"
		}
	}
	return strings.TrimSuffix(str, "|")
}

func (v EventType) FlagString() string {
	switch v {
	case EVENT_TYPE_NONE:
		return "EVENT_TYPE_NONE"
	case EVENT_TYPE_ADDED:
		return "EVENT_TYPE_ADDED"
	case EVENT_TYPE_REMOVED:
		return "EVENT_TYPE_REMOVED"
	case EVENT_TYPE_RENAMED:
		return "EVENT_TYPE_RENAMED"
	case EVENT_TYPE_CHANGED:
		return "EVENT_TYPE_CHANGED"
	default:
		return "[?? Invalid EventType]"
	}
}

package main

import (
	"io/fs"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type EventType int

type Event struct {
	Path string
	Info fs.FileInfo
	Type EventType
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	notifyChannelSize = 10
)

const (
	EventTypeNone EventType = iota
	EventTypeAdded
	EventTypeRemoved
	EventTypeRenamed
	EventTypeChanged
)

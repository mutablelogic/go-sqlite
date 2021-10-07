package indexer

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sync"
	// Package imports
	// Import namepaces
	//. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Queue struct {
	sync.RWMutex
	q []string
	k map[string]*QueueEvent
}

type QueueEvent struct {
	Event
	Name string
	Path string
	Info fs.FileInfo
}

type Event uint

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultQueueCapacity = 1024
)

const (
	EventNone Event = iota
	EventAdd
	EventRemove
	EventReindexStarted
	EventReindexCompleted
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new queue with default capacity
func NewQueue() *Queue {
	return NewQueueWithCapacity(0)
}

// Create a new queue which acts as a buffer between the file indexing
// and the processng/rendering which can be slower than the file indexing
func NewQueueWithCapacity(cap int) *Queue {
	q := new(Queue)

	// Create capacity
	if cap == 0 {
		q.q = make([]string, 0, defaultQueueCapacity)
		q.k = make(map[string]*QueueEvent, defaultQueueCapacity)
	} else {
		q.q = make([]string, 0, cap)
		q.k = make(map[string]*QueueEvent, cap)
	}

	// Return success
	return q
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Queue) String() string {
	str := "<queue"
	str += fmt.Sprint(" count=", this.Count())
	str += fmt.Sprintf(" cap=%q", this.q)
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Indicate reindexing in progress or completed
func (q *Queue) Mark(name, path string, flag bool) {
	if flag {
		q.add(EventReindexStarted, name, path, nil)
	} else {
		q.add(EventReindexCompleted, name, path, nil)
	}
}

// Add an item to the queue. If the item is already in the queue,
// then it is bumped to the end of the queue
func (q *Queue) Add(name, path string, info fs.FileInfo) {
	if elem := q.Get(name, path); elem != nil {
		// Remove the element from the existing queue
		q.del(name, path)
	}

	// Add the element to the queue
	q.add(EventAdd, name, path, info)
}

// Remove an item to the queue. If the item is already in the queue,
// it is removed
func (q *Queue) Remove(name, path string) {
	if elem := q.Get(name, path); elem != nil {
		// Remove the element from the existing queue
		q.del(name, path)
	}

	// Add the element to the queue
	q.add(EventRemove, name, path, nil)
}

// Return a queue event from the queue, or nil
func (q *Queue) Get(name, path string) *QueueEvent {
	q.RWMutex.RLock()
	defer q.RWMutex.RUnlock()
	if elem, exists := q.k[key(name, path)]; exists {
		return elem
	} else {
		return nil
	}
}

// Return a queue event from the head
func (q *Queue) Next() *QueueEvent {
	q.RWMutex.Lock()
	defer q.RWMutex.Unlock()
	for i, k := range q.q {
		if k != "" {
			if elem, exists := q.k[k]; exists {
				q.q[i] = ""
				delete(q.k, k)
				return elem
			}
		}
	}
	// No elements remain - garbarge collect
	q.q = q.q[:0]
	return nil
}

// Return a queue event from the head
func (q *Queue) Count() int {
	q.RWMutex.RLock()
	defer q.RWMutex.RUnlock()
	return len(q.k)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (q *Queue) add(e Event, name, path string, info fs.FileInfo) {
	q.RWMutex.Lock()
	defer q.RWMutex.Unlock()
	// This assumes the key does not exist
	key := key(name, path)
	if _, exists := q.k[key]; exists {
		panic("Queue: key already exists")
	}
	q.q = append(q.q, key)
	q.k[key] = &QueueEvent{e, name, path, info}
}

func (q *Queue) del(name, path string) {
	q.RWMutex.Lock()
	defer q.RWMutex.Unlock()
	key := key(name, path)
	// Set the key to empty in the array
	for i, k := range q.q {
		if k == key {
			q.q[i] = ""
			break
		}
	}
	// Remove from hash
	delete(q.k, key)
}

func key(name, path string) string {
	return filepath.Join(name, path)
}

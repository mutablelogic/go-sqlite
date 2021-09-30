package indexer

import (
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
	q []QueueEvent
	k map[string]*QueueEvent
}

type QueueEvent struct {
	name string
	path string
	info fs.FileInfo
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultQueueCapacity = 1024
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new queue which acts as a buffer between the file indexing
func NewQueue() *Queue {
	return NewQueueWithCapacity(defaultQueueCapacity)
}

// Create a new queue which acts as a buffer between the file indexing
// and the rendering which can be slower than the file indexing
func NewQueueWithCapacity(cap int) *Queue {
	q := new(Queue)

	// Create capacity
	if cap == 0 {
		q.q = make([]QueueEvent, 0, defaultQueueCapacity)
		q.k = make(map[string]*QueueEvent, defaultQueueCapacity)
	} else {
		q.q = make([]QueueEvent, 0, cap)
		q.k = make(map[string]*QueueEvent, cap)
	}

	// Return success
	return q
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Add an item to the queue. If the item is already in the queue,
// then it is bumped to the end of the queue
func (q *Queue) Add(name, path string) error {
	/*	if elem := q.Get(name, path); elem != nil {
			// Remove the element from the existing queue
			q.pop(elem)
		}
		// Add the element to the queue
		q.push(elem)*/
	return nil
}

// Remove an item to the queue. If the item is already in the queue,
// it is removed
func (q *Queue) Remove(name, path string) error {
	/*	if elem := q.Get(name, path); elem != nil {
			// Remove the element from the existing queue
			q.pop(elem)
		} else {
			return ErrNotFound.With(name, path)
		}*/
	// Return success
	return nil
}

// Return a queue event from the queue, or nil
func (q *Queue) Get(name, path string) *QueueEvent {
	q.RWMutex.RLock()
	defer q.RWMutex.RUnlock()
	key := filepath.Join(name, path)
	if elem, exists := q.k[key]; exists {
		return elem
	} else {
		return nil
	}
}

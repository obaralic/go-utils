// -----------------------------------------------------------------------------
// The purpose of the pool package is to show how you can use a buffered channel
// to pool a set of resources that can be shared and individually used by any
// number of goroutines. This pattern is useful when you have a static set of
// resources to share, such as database connections or memory buffers.
// When a goroutine needs one of these resources from the pool,
// it can acquire the resource, use it, and then return it to the pool.
// -----------------------------------------------------------------------------
package pool

import (
	"errors"
	"io"
	"sync"
)

// -----------------------------------------------------------------------------
// Pool manages a set of resources safely shared by multiple goroutines.
// The resource being managed must implement the io.Closer interface.
// -----------------------------------------------------------------------------
type Pool struct {
	mutex     sync.Mutex
	resources chan io.Closer
	factory   func() (io.Closer, error)
	closed    bool
}

// ErrorPoolClosed - returned when an Acquire returns on a closed pool.
var ErrorPoolClosed = errors.New("Pool has been closed.")

// -----------------------------------------------------------------------------
// New - Creates a pool that manages resources.
// A pool requires a function that can allocate a new resource
// and the size of the pool.
// -----------------------------------------------------------------------------
func New(allocator func() (io.Closer, error), size uint) (*Pool, error) {
	if size <= 0 {
		return nil, errors.New("Pool size is to small.")
	}

	return &Pool{
		resources: make(chan io.Closer, size),
		factory:   allocator,
	}, nil
}

// -----------------------------------------------------------------------------
// Acquire - Retrieves a resource	from the pool.
// -----------------------------------------------------------------------------
func (pool *Pool) Acquire() (io.Closer, error) {
	select {
	// Check for a free resource.
	case resource, ok := <-pool.resources:
		if !ok {
			return nil, ErrorPoolClosed
		}
		return resource, nil

	// Provide a new resource since there are none available.
	default:
		return pool.factory()
	}
}

// -----------------------------------------------------------------------------
// Release - Places a new resource into the pool.
// -----------------------------------------------------------------------------
func (pool *Pool) Release(resource io.Closer) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	if pool.closed {
		resource.Close()
		return
	}

	select {
	// Attempt to place the new resource on the queue.
	case pool.resources <- resource:

	// If the queue is already at cap we close the resource.
	default:
		resource.Close()
	}
}

// -----------------------------------------------------------------------------
// Close - Shutdown the pool and close all existing resources.
// -----------------------------------------------------------------------------
func (pool *Pool) Close() {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	if pool.closed {
		return
	} else {
		pool.closed = true
	}

	// Close the channel before we drain the channel of its resources.
	// If we don't do this, we will have a deadlock.
	close(pool.resources)

	for resource := range pool.resources {
		resource.Close()
	}
}

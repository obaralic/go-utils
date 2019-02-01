// -----------------------------------------------------------------------------
// The purpose of the work package is to show how you can use an unbuffered
// channel to create a pool of goroutines that will perform and control
// the amount of work that gets done concurrently.
// This is a better approach than using a buffered channel of some arbitrary
// static size that acts as a queue of work and throwing goroutines at it.
// Unbuffered channels provide a guarantee that data has been exchanged between
// two goroutines. This approach of using an unbuffered channel allows the user
// to know when the pool is performing the work, and the channel pushes back
// when it can’t accept any more work because it’s busy.
// No work is ever lost or stuck in a queue
// that has no guarantee it will ever be worked on.
// -----------------------------------------------------------------------------
package work

import (
	"errors"
	"sync"
)

// -----------------------------------------------------------------------------
// Worker - Interface that must be implemented by the client of the work pool.
// -----------------------------------------------------------------------------
type Worker interface {
	Work()
}

// -----------------------------------------------------------------------------
// WokrPool - Provides pool of goroutines that can execute submitted work.
// -----------------------------------------------------------------------------
type WorkPool struct {
	workers chan Worker
	barrier sync.WaitGroup
}

// -----------------------------------------------------------------------------
// New - Creates a new worker pool that waits to the work to be submitted.
// -----------------------------------------------------------------------------
func New(size int) (*WorkPool, error) {
	if size <= 0 {
		return nil, errors.New("Pool size is to small.")
	}

	pool := WorkPool{
		workers: make(chan Worker),
	}

	pool.barrier.Add(size)

	for c := 0; c < size; c++ {
		go func() {
			for worker := range pool.workers {
				worker.Work()
			}
			pool.barrier.Done()
		}()
	}

	return &pool, nil
}

// -----------------------------------------------------------------------------
// Submit - Submits work to the pool.
// -----------------------------------------------------------------------------
func (pool *WorkPool) Submit(worker Worker) {
	pool.workers <- worker
}

// -----------------------------------------------------------------------------
// Close - Waits for all the goroutines to shutdown.
// -----------------------------------------------------------------------------
func (pool *WorkPool) Close() {
	close(pool.workers)
	pool.barrier.Wait()
}

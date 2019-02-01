// -----------------------------------------------------------------------------
// The purpose of the runner package is to show how channels can be used
// to monitor the amount of time a program is running and terminate the program
// if it runs too long. This pattern is useful when developing a program
// that will be scheduled to run as a background task process.
// This could be a program that runs as a cron job,
// or in a worker-based cloud environment like Iron.io.
// -----------git rm -r --cached .------------------------------------------------------------------
package runner

import (
	"errors"
	"os"
	"os/signal"
	"time"
)

// -----------------------------------------------------------------------------
// Runner runs a set of tasks within a given timeout,
// and can be shut down on a OS interrupt.
// -----------------------------------------------------------------------------
type Runner struct {

	// Interrupt channel - reports a signal from the OS.
	interrupt chan os.Signal

	// Complete channel - reports that process is done.
	complete chan error

	// Timeout channel - reports that time has run out.
	timeout <-chan time.Time

	// Tasks - functions that are executed synchronously.
	tasks []func(int)
}

// ErrorTimeout - returned when a value is received on the timeout.
var ErrorTimeout = errors.New("Timeout received")

// ErrorInterrupt - returned when a value is received on the interrupt.
var ErrorInterrupt = errors.New("Interrupt received")

// -----------------------------------------------------------------------------
// New - constructor pattern that returns ready to run Runner.
// -----------------------------------------------------------------------------
func New(duration time.Duration) *Runner {
	return &Runner{
		interrupt: make(chan os.Signal, 1),
		complete:  make(chan error),
		timeout:   time.After(duration),
	}
}

// -----------------------------------------------------------------------------
// Add - attaches tasks to the Runner.
// Task is a function that takes an int ID.
// -----------------------------------------------------------------------------
func (runner *Runner) Add(tasks ...func(int)) {
	runner.tasks = append(runner.tasks, tasks...)
}

// -----------------------------------------------------------------------------
// Start - runs all tasks and monitors channel events.
// -----------------------------------------------------------------------------
func (runner *Runner) Start() error {
	// We want to receive all interrupt based signals.
	signal.Notify(runner.interrupt, os.Interrupt)

	go func() {
		runner.complete <- runner.run()
	}()

	select {
	case error := <-runner.complete:
		return error

	case <-runner.timeout:
		return ErrorTimeout
	}
}

// -----------------------------------------------------------------------------
// run - executes each registered task.
// -----------------------------------------------------------------------------
func (runner *Runner) run() error {
	for id, task := range runner.tasks {
		if runner.interrupted() {
			return ErrorInterrupt
		}
		task(id)
	}
	return nil
}

// -----------------------------------------------------------------------------
// interrupted - verifies if the interrupt signal has been issued.
// -----------------------------------------------------------------------------
func (runner *Runner) interrupted() bool {
	// Lets a goroutine wait on multiple communication operations.
	select {
	case <-runner.interrupt:
		// Stop receiving any further signals.
		signal.Stop(runner.interrupt)
		return true

	default:
		return false
	}
}

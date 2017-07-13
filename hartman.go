// Package hartman supervises and keeps n workers at all times. https://en.wikipedia.org/wiki/Full_Metal_Jacket
package hartman

import (
	"context"
	"errors"
	"log"
	"sync"
)

// ErrAllDone lets supervisor that to complete a worker
var ErrAllDone = errors.New("no more work")

// Worker defines Work func
type Worker interface {
	Work(ctx context.Context) error
}

// Supervisor supervises the workers
type Supervisor struct {
	sync.Mutex
	errors chan error

	Ctx             context.Context
	NumberOfWorkers int // number of workers
	Worker          Worker
	ErrorHandler    func(errors <-chan error)
	DoneHandler     func()
}

// NewSupervisor creates Supervisor
func NewSupervisor(ctx context.Context, numWorkers int, worker Worker) *Supervisor {
	return &Supervisor{
		Ctx:             ctx,
		NumberOfWorkers: numWorkers,
		Worker:          worker,
	}
}

// SetErrorHandler sets a error handler and runs it in a routine
func (s *Supervisor) SetErrorHandler(errHandler func(errors <-chan error)) {
	s.Lock()
	defer s.Unlock()

	s.ErrorHandler = errHandler
}

// SetDoneHandler sets a done handler func which is called when supervision is done
func (s *Supervisor) SetDoneHandler(doneHandler func()) {
	s.Lock()
	defer s.Unlock()

	s.DoneHandler = doneHandler
}

// Work implements Worker
func (s *Supervisor) Work(ctx context.Context) error {
	s.errors = make(chan error)
	defer close(s.errors)

	if s.ErrorHandler == nil {
		s.ErrorHandler = func(errors <-chan error) {
			for err := range errors {
				log.Printf("supervisor received error with: %v", err)
			}
		}
	}

	go s.ErrorHandler(s.errors)

	var wg sync.WaitGroup
	for i := 0; i < s.NumberOfWorkers; i++ {
		wg.Add(1)
		// start workers
		go func(ctx context.Context) {
			defer wg.Done()
			for ctx.Err() == nil {
				switch err := s.Worker.Work(ctx); err {
				case nil:
					/* nop */
				case ErrAllDone:
					return
				default:
					s.errors <- err
				}
			}
		}(ctx)
	}
	wg.Wait()

	if s.DoneHandler != nil {
		s.DoneHandler()
	}

	return nil
}

// RunC runs fns concurrently, finishes if any worker errors
func RunC(ctx context.Context, fns ...Worker) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errors := make(chan error)
	var wg sync.WaitGroup

	go func() {
		wg.Wait()
		close(errors)
	}()

	for _, fn := range fns {
		wg.Add(1)
		go func(fn Worker) {
			defer wg.Done()
			err := fn.Work(ctx)

			if err != nil {
				errors <- err
			}
		}(fn)
	}

	for err := range errors {
		return err
	}

	return nil
}

// RunF runs fns concurrently, finishes if any worker successfully completes
func RunF(ctx context.Context, fns ...Worker) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	done := make(chan error)
	var wg sync.WaitGroup

	go func() {
		wg.Wait()
		close(done)
	}()

	for _, fn := range fns {
		wg.Add(1)
		go func(fn Worker) {
			defer wg.Done()
			err := fn.Work(ctx)

			if err == nil {
				done <- nil
			}
		}(fn)
	}

	return <-done
}

// RunS runs fns concurrently, restarts any fn that completes untill ctx is cancelled
func RunS(ctx context.Context, fns ...Worker) <-chan error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errors := make(chan error)
	defer close(errors)
	var wg sync.WaitGroup
	go func() {
		wg.Wait()
	}()

	for _, fn := range fns {
		wg.Add(1)
		go func(fn Worker) {
			defer wg.Done()
			for ctx.Err() == nil {
				err := fn.Work(ctx)
				if err != nil {
					errors <- err
				}
			}
		}(fn)
	}

	return errors
}

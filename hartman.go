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

// Supervise starts the workers
func (s *Supervisor) Supervise() {
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
	wg.Add(s.NumberOfWorkers)

	for i := 0; i < s.NumberOfWorkers; i++ {
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
		}(s.Ctx)
	}

	wg.Wait()
	if s.DoneHandler != nil {
		s.DoneHandler()
	}
}

// RunGroup runs fns concurrently, finishes if any worker errors
func RunGroup(ctx context.Context, fns ...Worker) error {
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
		log.Printf("not able to run worker, failed with: %v", err)

		return err
	}

	return nil
}

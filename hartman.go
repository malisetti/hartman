// Package hartman.
// https://en.wikipedia.org/wiki/Full_Metal_Jacket
package hartman

import (
	"context"
	"sync"
)

// Work is func
type Work = func(ctx context.Context) error

// RunC runs fns concurrently, finishes if any worker errors
func RunC(ctx context.Context, fns ...Work) error {
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
		go func(fn Work) {
			defer wg.Done()
			err := fn(ctx)

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
func RunF(ctx context.Context, fns ...Work) error {
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
		go func(fn Work) {
			defer wg.Done()
			err := fn(ctx)

			if err == nil {
				errors <- nil
			}
		}(fn)
	}

	return <-errors
}

// RunN runs given fn concurrently on n go routines, finishes if any error from a routine or when all completes
func RunN(ctx context.Context, fn Work, n int) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errors := make(chan error)
	defer close(errors)

	var wg sync.WaitGroup
	go func() {
		wg.Wait()
		close(errors)
	}()

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(ctx context.Context) {
			defer wg.Done()
			err := fn(ctx)
			if err != nil {
				errors <- err
				cancel()
			}
		}(ctx)
	}

	return <-errors
}

// RunS runs fns concurrently, restarts any fn that completes untill ctx is cancelled
func RunS(ctx context.Context, fns ...Work) <-chan error {
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
		go func(fn Work) {
			defer wg.Done()
			for ctx.Err() == nil {
				err := fn(ctx)
				if err != nil {
					errors <- err
				}
			}
		}(fn)
	}

	return errors
}

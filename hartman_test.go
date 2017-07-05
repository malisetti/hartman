package hartman

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"testing"
	"time"
)

type LazyWorker struct {
}

func (w *LazyWorker) Work(ctx context.Context) error {
	timer := time.NewTimer(1 * time.Second)
	for {
		select {
		case <-time.After(2 * time.Second):
			return fmt.Errorf("Done my work")
		case <-ctx.Done():
			return nil
		case <-timer.C:
			time.Sleep(1 * time.Second)
		}
	}
}

func TestSupervise(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(20*time.Second, func() {
		cancel()
	})
	numWorkers := 2
	s := NewSupervisor(ctx, numWorkers, &LazyWorker{})

	s.SetDoneHandler(func() {
		log.Println("Done supervising")
	})

	s.SetErrorHandler(func(errors <-chan error) {
		for err := range errors {
			log.Printf("in error handler %v\n", err)
		}
	})

	s.Supervise()
}

type LazyWorker2 struct {
	WorkChan chan string
}

func (w *LazyWorker2) Work(ctx context.Context) error {
	for {
		select {
		case s, ok := <-w.WorkChan:
			if !ok {
				return ErrAllDone
			}

			time.Sleep(2 * time.Second)
			log.Println("got work", s)
		case <-ctx.Done():
			return nil
		}
	}
}

func TestSupervise2(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(20*time.Second, func() {
		cancel()
	})
	numWorkers := 2
	workChan := make(chan string)
	go func() {
		defer close(workChan)
		for i := 0; i < 20; i++ {
			workChan <- strconv.Itoa(i)
		}
	}()

	s := NewSupervisor(ctx, numWorkers, &LazyWorker2{WorkChan: workChan})

	s.SetDoneHandler(func() {
		log.Println("Done supervising")
	})

	s.SetErrorHandler(func(errors <-chan error) {
		for err := range errors {
			log.Printf("in error handler %v\n", err)
		}
	})

	s.Supervise()
}

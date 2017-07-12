package hartman

import (
	"context"
	"fmt"
	"log"
	"math/rand"
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
	ti := time.AfterFunc(20*time.Second, func() {
		cancel()
	})
	defer ti.Stop()

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
	ti := time.AfterFunc(20*time.Second, func() {
		cancel()
	})
	defer ti.Stop()

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

/**
	Will try make sandwich
	need tomoto shopper
	need cheese shopper
	need waiter
**/

type TShopper struct {
}

type CShopper struct {
}

type SWaiter struct {
}

func (w *TShopper) Work(ctx context.Context) error {
	select {
	case <-time.After(5 * time.Second):
		log.Println("shopped tomatoes")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("could not shop tomatoes")
	}
}

func (w *CShopper) Work(ctx context.Context) error {
	select {
	case <-time.After(4 * time.Second):
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		if r.Intn(100)%2 == 0 {
			return fmt.Errorf("cheese is not available. sorry, no sanwich")
		}
		log.Println("shopped cheese")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("could not shop cheese")
	}
}

func (w *SWaiter) Work(ctx context.Context) error {
	select {
	case <-time.After(10 * time.Second):
		log.Printf("sandwich is ready")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("order cancelled")
	}
}

func TestRunGroup(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	err := RunGroup(ctx, &TShopper{}, &CShopper{}, &SWaiter{})

	if err != nil {
		t.Error(err)
	}
}

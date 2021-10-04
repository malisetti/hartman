package hartman

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"testing"
	"time"
)

/**
	Will try make sandwich
	need tomoto shopper
	need cheese shopper
	need waiter
**/

func TShopper(ctx context.Context) error {
	select {
	case <-time.After(5 * time.Second):
		log.Println("shopped tomatoes")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("could not shop tomatoes")
	}
}

func CShopper(ctx context.Context) error {
	select {
	case <-time.After(4 * time.Second):
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		if r.Intn(100)%2 == 0 {
			return fmt.Errorf("cheese is not available. sorry, no sandwich")
		}
		log.Println("shopped cheese")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("could not shop cheese")
	}
}

func SWaiter(ctx context.Context) error {
	select {
	case <-time.After(10 * time.Second):
		log.Printf("sandwich is ready")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("order cancelled")
	}
}

func TestRunC(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	err := RunC(ctx, TShopper, CShopper, SWaiter)

	if err != nil {
		t.Error(err)
	}
}

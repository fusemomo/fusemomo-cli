package tui

import (
	"fmt"
	"os"
	"time"
)

// RunBatchWithProgress runs fn for a batch operation and shows progress on stderr.
func RunBatchWithProgress[T any](fn func() (T, error)) (T, error) {
	if !isStderrTTY() {
		return fn()
	}

	type result struct {
		val T
		err error
	}
	ch := make(chan result, 1)
	start := time.Now()

	// Show spinner while batch is in flight.
	done := make(chan struct{})
	go func() {
		frames := spinnerFrames
		i := 0
		for {
			select {
			case <-done:
				return
			default:
				fmt.Fprintf(os.Stderr, "\r%s Processing batch...", frames[i%len(frames)])
				i++
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	go func() {
		v, err := fn()
		ch <- result{val: v, err: err}
	}()

	res := <-ch
	close(done)
	elapsed := time.Since(start)

	// Clear the spinner line.
	fmt.Fprintf(os.Stderr, "\r")

	if res.err == nil {
		if r, ok := any(res.val).(interface{ GetLoggedCount() int }); ok {
			fmt.Fprintf(os.Stderr, "Logged %d interactions in %.1fs\n", r.GetLoggedCount(), elapsed.Seconds())
		} else {
			fmt.Fprintf(os.Stderr, "Batch completed in %.1fs\n", elapsed.Seconds())
		}
	}

	return res.val, res.err
}

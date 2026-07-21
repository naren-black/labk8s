// Package concurrency is Lesson 4's exercise: goroutines, channels, and the
// concurrency patterns you'll see constantly in real Kubernetes client code
// (informers, workqueues, calling multiple API replicas). Read
// ../tour/main.go first (just run it).
//
// Hint: after you get these passing with `go test`, also run:
//
//	go test -race ./00-go-refresher/04-concurrency/exercise/...
//
// The race detector catches bugs these tests can't - e.g. returning a map
// that a background goroutine might still write to.
package concurrency

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// ErrAPINotFound is a sentinel (see Lesson 3) returned by simulateAPICall
// for negative ids.
var ErrAPINotFound = errors.New("API not found")

// ErrTimeout is a sentinel returned when a fetch doesn't complete in time.
var ErrTimeout = errors.New("timed out waiting for API calls")

// simulateAPICall pretends to call an external service for the given id.
// Real APIs have latency - we simulate it with a sleep proportional to id
// (id=1 -> 20ms, id=2 -> 40ms, ...) so behavior is deterministic for tests.
// Negative ids simulate a "not found" failure with no latency.
func simulateAPICall(id int) (string, error) {
	if id < 0 {
		return "", fmt.Errorf("id %d: %w", id, ErrAPINotFound)
	}
	time.Sleep(time.Duration(id) * 20 * time.Millisecond)
	return fmt.Sprintf("response-%d", id), nil
}

// FetchAll calls simulateAPICall for every id CONCURRENTLY (not one at a
// time - that's the whole point) and collects the results.
//
// TODO: implement.
//   - Launch one goroutine per id.
//   - Protect the shared results map with a sync.Mutex (or send results
//     over a channel and collect in the main goroutine - your choice).
//   - Use a sync.WaitGroup so you know when every goroutine is done.
//   - If simulateAPICall returns an error for some id, that id is simply
//     left out of the results map, but its error should still be reported.
//   - Combine multiple errors with errors.Join(errs...) - it's like %w but
//     for combining SEVERAL errors into one. errors.Is/errors.As still work
//     through it: errors.Is(joined, ErrAPINotFound) is true if ANY of the
//     joined errors matches.
func FetchAll(ids []int) (map[int]string, error) {
	results := make(map[int]string)
	errs := []error{}
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	for _, id := range ids {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			response, err := simulateAPICall(id)
			if err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			} else {
				mu.Lock()
				results[id] = response
				mu.Unlock()
			}
		}(id)
	}
	wg.Wait()
	return results, errors.Join(errs...)
}

// FetchAllWithTimeout is like FetchAll, but gives up after timeout and
// returns whatever results were collected so far, plus ErrTimeout.
//
// TODO: implement.
//   - Same fan-out as FetchAll, but also start a goroutine that waits on
//     your sync.WaitGroup and then closes a `done chan struct{}`.
//   - select on `done` vs time.After(timeout).
//   - If done fires first: return the full results and a nil error (or
//     joined per-id errors, same as FetchAll).
//   - If the timeout fires first: return a COPY of whatever results exist
//     so far (copy the map while holding the mutex, so you're not handing
//     back a map a slower goroutine might still write to) and an error
//     that satisfies errors.Is(err, ErrTimeout).
//   - IMPORTANT: goroutines you launched but didn't wait for keep running
//     in the background after you return - they just write to a map
//     nobody looks at anymore. That's a real limitation of this simple
//     version; real code cancels them properly with context.Context
//     (a later lesson).
func FetchAllWithTimeout(ids []int, timeout time.Duration) (map[int]string, error) {
	results := make(map[int]string)
	errs := []error{}
	wg := sync.WaitGroup{}
	done := make(chan struct{})
	mu := sync.Mutex{}
	for _, id := range ids {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			response, err := simulateAPICall(id)
			if err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			} else {
				mu.Lock()
				results[id] = response
				mu.Unlock()
			}
		}(id)
	}
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return results, errors.Join(errs...)
	case <-time.After(timeout):
		mu.Lock()
		resultsCopy := make(map[int]string)
		for k, v := range results {
			resultsCopy[k] = v
		}
		mu.Unlock()
		return resultsCopy, ErrTimeout
	}
}

// FirstSuccessful calls simulateAPICall for every id concurrently (think:
// hitting several redundant replicas of the same API) and returns the FIRST
// one that succeeds, ignoring the rest. If all of them fail, return a
// joined error containing every failure.
//
// TODO: implement.
//   - Use a channel sized len(ids) (buffered!) so goroutines that "lose the
//     race" can still send their result without blocking forever waiting
//     for a receiver that will never come.
//   - Receive from the channel in a loop: return immediately on the first
//     nil-error result.
//   - If you drain all len(ids) results without finding a success, return
//     "", errors.Join(all the errors).
func FirstSuccessful(ids []int) (string, error) {
	type apiResult struct {
		response string
		err error
	}
	var errs []error
	results := make(chan apiResult, len(ids))
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	for _, id := range ids {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			response, err := simulateAPICall(id)
			if err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			} else {
				mu.Lock()
				results <- apiResult{response: response, err: nil}
				mu.Unlock()
			}
		}(id)
	}
	wg.Wait()
	close(results)
	for result := range results {
		if result.err == nil {
			return result.response, nil
		}
		errs = append(errs, result.err)
	}
	return "", errors.Join(errs...) // join all errors
}


// ProcessJobs is a worker pool: exactly numWorkers goroutines pull job
// numbers off a shared channel and compute job*job, until the channel is
// closed and drained. This is the same shape client-go's workqueue uses to
// fan reconciliation work out across a fixed number of goroutines.
//
// TODO: implement.
//   - Make a channel, send every job into it, then close it (only the
//     sender closes - see the tour).
//   - Launch numWorkers goroutines, each ranging over the jobs channel.
//   - Protect the shared results map with a sync.Mutex, or use sync.Map.
//   - Use a sync.WaitGroup to know when all workers are done before you
//     return the results.
func ProcessJobs(jobs []int, numWorkers int) map[int]int {
	results := make(map[int]int)
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	jobsChan := make(chan int, len(jobs))
	for _, job := range jobs {
		jobsChan <- job
	}
	close(jobsChan)
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for job := range jobsChan {
				mu.Lock()
				results[job] = job * job
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()
	return results
}
// Narrated tour, nothing to fill in. Run with:
//
//	go run ./00-go-refresher/04-concurrency/tour
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	section("1. A goroutine without synchronization is a race, not a guarantee")
	go fmt.Println("hello from a goroutine")
	fmt.Println("hello from main")
	time.Sleep(50 * time.Millisecond)
	// The `go` statement schedules a function to run concurrently and returns
	// immediately. Without the Sleep, main could exit before the goroutine
	// ever runs. Sleep-to-synchronize is a hack - section 2 shows the real fix.

	section("2. sync.WaitGroup: wait for N goroutines to finish")
	var wg sync.WaitGroup
	results := make([]int, 5)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i] = i * i
			// NOTE: this closure captures `i` by reference to the loop
			// variable. On Go 1.22+ (this module targets 1.26), each loop
			// iteration gets its OWN `i`, so this is safe and prints
			// 0 1 4 9 16 in some order. Before 1.22, every closure shared
			// ONE `i`, and by the time goroutines ran, i was usually 5 for
			// all of them - a classic bug. Worth knowing if you ever read
			// pre-2023 Go code or run an older toolchain.
		}()
	}
	wg.Wait()
	fmt.Println("results =", results)

	section("3. Unbuffered channel: send blocks until someone receives")
	ch := make(chan string)
	go func() {
		fmt.Println("goroutine: about to send")
		ch <- "payload"
		fmt.Println("goroutine: send completed")
	}()
	time.Sleep(50 * time.Millisecond)
	fmt.Println("main: about to receive")
	msg := <-ch
	fmt.Println("main: received", msg)
	// The goroutine's send BLOCKS at `ch <- "payload"` until main's `<-ch`
	// runs. An unbuffered channel is a synchronization point, not a queue.

	section("4. Buffered channel: send doesn't block until the buffer is full")
	buffered := make(chan int, 2)
	buffered <- 1
	buffered <- 2
	fmt.Println("sent two values without any receiver - buffer holds 2")
	fmt.Println("received:", <-buffered, <-buffered)

	section("5. Closing a channel + range: signal 'no more values'")
	jobs := make(chan int)
	go func() {
		for i := 1; i <= 3; i++ {
			jobs <- i
		}
		close(jobs)
		// close() tells receivers no more values are coming. Only the
		// SENDER should close a channel, never the receiver.
	}()
	for j := range jobs {
		// range over a channel receives values until it's closed AND drained.
		fmt.Println("received job", j)
	}

	section("6. select: wait on whichever channel is ready first")
	c1 := make(chan string)
	c2 := make(chan string)
	go func() {
		time.Sleep(30 * time.Millisecond)
		c1 <- "from c1"
	}()
	go func() {
		time.Sleep(10 * time.Millisecond)
		c2 <- "from c2"
	}()
	for i := 0; i < 2; i++ {
		select {
		case v := <-c1:
			fmt.Println("select got:", v)
		case v := <-c2:
			fmt.Println("select got:", v)
		}
	}
	// select picks whichever `case` is ready. c2 fires first here (10ms vs
	// 30ms sleep), so it's printed first even though c1's case is listed
	// first in the source.

	section("7. select with a timeout, using time.After")
	slow := make(chan string)
	select {
	case v := <-slow:
		fmt.Println("got:", v)
	case <-time.After(50 * time.Millisecond):
		fmt.Println("timed out waiting for slow")
		// time.After returns a channel that fires once after the given
		// duration. Racing it against your real channel in a select is the
		// standard Go pattern for "wait, but not forever".
	}

	section("8. sync.Mutex: maps and shared state need explicit locking")
	counts := map[string]int{}
	var mu sync.Mutex
	var wg2 sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			mu.Lock()
			counts["events"]++
			mu.Unlock()
			// Without the mutex this is a DATA RACE: concurrent map writes
			// from multiple goroutines can corrupt the map or panic with
			// "fatal error: concurrent map writes". Try running with
			// `go run -race ./00-go-refresher/04-concurrency/tour` after
			// commenting out the Lock/Unlock lines to see it for real.
		}()
	}
	wg2.Wait()
	fmt.Println("counts[events] =", counts["events"], "(want 100)")

	section("9. Worker pool: fixed number of goroutines pulling from one channel")
	workQueue := make(chan int, 10)
	var results2 sync.Map // concurrent-safe map, an alternative to mutex+map
	var wg3 sync.WaitGroup
	for w := 1; w <= 3; w++ {
		wg3.Add(1)
		go func(workerID int) {
			defer wg3.Done()
			for job := range workQueue {
				results2.Store(job, job*job)
				_ = workerID // in real code you'd likely log this
			}
			// Each worker ranges over the SAME channel. Go's channel
			// semantics guarantee each value is delivered to exactly one
			// receiver - this is how Kubernetes controllers fan work out
			// across a fixed pool of goroutines (see client-go's workqueue).
		}(w)
	}
	for j := 1; j <= 10; j++ {
		workQueue <- j
	}
	close(workQueue)
	wg3.Wait()
	v, _ := results2.Load(7)
	fmt.Println("results2[7] =", v, "(want 49)")
}

func section(title string) {
	fmt.Println()
	fmt.Println("===", title, "===")
}

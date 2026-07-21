package concurrency

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestFetchAll(t *testing.T) {
	got, err := FetchAll([]int{1, 2, 3, -1, -2})
	if len(got) != 3 {
		t.Fatalf("got %d results, want 3 (ids 1,2,3 should succeed): %v", len(got), got)
	}
	for _, id := range []int{1, 2, 3} {
		want := fmt.Sprintf("response-%d", id)
		if got[id] != want {
			t.Fatalf("got[%d]=%q, want %q", id, got[id], want)
		}
	}
	if err == nil {
		t.Fatalf("expected a non-nil error (ids -1,-2 fail), got nil")
	}
	if !errors.Is(err, ErrAPINotFound) {
		t.Fatalf("errors.Is(err, ErrAPINotFound) = false, want true")
	}
}

func TestFetchAllWithTimeout(t *testing.T) {
	start := time.Now()
	got, err := FetchAllWithTimeout([]int{1, 2, 3, 4}, 45*time.Millisecond)
	elapsed := time.Since(start)

	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("errors.Is(err, ErrTimeout) = false, want true")
	}
	if got[1] != "response-1" || got[2] != "response-2" {
		t.Fatalf("got %v, want ids 1 and 2 present (20ms, 40ms < 45ms timeout)", got)
	}
	if _, ok := got[3]; ok {
		t.Fatalf("got id 3 in results, want it absent (60ms > 45ms timeout)")
	}
	if elapsed > 300*time.Millisecond {
		t.Fatalf("took %v, want well under 300ms (should return at the timeout, not wait for slow ids)", elapsed)
	}
}

func TestFirstSuccessful(t *testing.T) {
	got, err := FirstSuccessful([]int{1, 2, 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "response-1" {
		t.Fatalf("got %q, want %q (id 1 is fastest at 20ms)", got, "response-1")
	}
}

func TestFirstSuccessful_AllFail(t *testing.T) {
	_, err := FirstSuccessful([]int{-1, -2, -3})
	if err == nil {
		t.Fatalf("expected an error when every id fails, got nil")
	}
	if !errors.Is(err, ErrAPINotFound) {
		t.Fatalf("errors.Is(err, ErrAPINotFound) = false, want true")
	}
}

func TestProcessJobs(t *testing.T) {
	got := ProcessJobs([]int{1, 2, 3, 4, 5}, 2)
	want := map[int]int{1: 1, 2: 4, 3: 9, 4: 16, 5: 25}
	if len(got) != len(want) {
		t.Fatalf("got %d results, want %d", len(got), len(want))
	}
	for job, square := range want {
		if got[job] != square {
			t.Fatalf("got[%d]=%d, want %d", job, got[job], square)
		}
	}
}

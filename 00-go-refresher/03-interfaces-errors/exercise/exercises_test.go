package interfaceserrors

import (
	"errors"
	"testing"
)

func newTestWorkload() *Workload {
	return &Workload{
		Name: "api",
		Containers: []ContainerDefinition{
			{Name: "app", Image: "myapp:v1"},
			{Name: "sidecar", Image: "envoy:v2"},
		},
	}
}

func TestFindContainer_Found(t *testing.T) {
	w := newTestWorkload()
	c, err := w.FindContainer("sidecar")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Image != "envoy:v2" {
		t.Fatalf("got Image=%q, want %q", c.Image, "envoy:v2")
	}
}

func TestFindContainer_NotFound(t *testing.T) {
	w := newTestWorkload()
	_, err := w.FindContainer("nope")
	if err == nil {
		t.Fatalf("expected an error, got nil")
	}
	if !errors.Is(err, ErrContainerNotFound) {
		t.Fatalf("errors.Is(err, ErrContainerNotFound) = false, want true (did you wrap with %%w?)")
	}
}

func TestSetContainerImage_Success(t *testing.T) {
	w := newTestWorkload()
	if err := w.SetContainerImage("app", "myapp:v2"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c, _ := w.FindContainer("app")
	if c.Image != "myapp:v2" {
		t.Fatalf("got Image=%q, want %q (mutation didn't stick - check index-based update)", c.Image, "myapp:v2")
	}
}

func TestSetContainerImage_NotFound(t *testing.T) {
	w := newTestWorkload()
	err := w.SetContainerImage("nope", "myapp:v2")
	if !errors.Is(err, ErrContainerNotFound) {
		t.Fatalf("errors.Is(err, ErrContainerNotFound) = false, want true")
	}
}

func TestSetContainerImage_EmptyImage(t *testing.T) {
	w := newTestWorkload()
	err := w.SetContainerImage("app", "")
	if err == nil {
		t.Fatalf("expected an error for empty image, got nil")
	}
	var ie *InvalidImageError
	if !errors.As(err, &ie) {
		t.Fatalf("errors.As(err, &InvalidImageError) = false, want true (did you wrap with %%w?)")
	}
	if ie.Container != "app" {
		t.Fatalf("got InvalidImageError.Container=%q, want %q", ie.Container, "app")
	}
}

func TestValidate(t *testing.T) {
	w := newTestWorkload()
	if err := w.Validate(); err != nil {
		t.Fatalf("expected valid workload, got error: %v", err)
	}

	empty := &Workload{Name: "empty"}
	if err := empty.Validate(); err == nil {
		t.Fatalf("expected error for workload with zero containers, got nil")
	}
}

func TestWorkloadSatisfiesValidator(t *testing.T) {
	// This is a COMPILE-TIME check: if *Workload doesn't have a Validate()
	// error method, this line fails to compile, not just fails at runtime.
	var v Validator = &Workload{Name: "x"}
	_ = v
}

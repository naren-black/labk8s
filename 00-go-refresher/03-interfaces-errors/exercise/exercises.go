// Package interfaceserrors is Lesson 3's exercise: interfaces and idiomatic
// error handling. Read ../tour/main.go first (just run it).
package interfaceserrors

import (
	"errors"
	"fmt"
)

// ErrContainerNotFound is a SENTINEL error - a package-level value compared
// by identity (with errors.Is), not by string message.
//
// TODO: declare it with errors.New("container not found").


var ErrContainerNotFound = errors.New("container not found")


// ContainerDefinition mirrors earlier lessons.
type ContainerDefinition struct {
	Name  string
	Image string
}

type Workload struct {
	Name       string
	Containers []ContainerDefinition
}

// FindContainer returns the container with the given name.
//
// TODO: implement. If not found, return a nil ContainerDefinition zero value
// and an error built by wrapping ErrContainerNotFound with context, e.g.:
//
//	fmt.Errorf("workload %s: %w", w.Name, ErrContainerNotFound)
//
// Use %w (not %v) so callers can still errors.Is() against ErrContainerNotFound
// even though your message adds extra context.
func (w *Workload) FindContainer(name string) (ContainerDefinition	, error) {
	for _, container := range w.Containers {
		if container.Name == name {
			return container, nil
		}
	}
	return ContainerDefinition{}, fmt.Errorf("workload %s: %w", w.Name, ErrContainerNotFound)
}

// InvalidImageError is a CUSTOM error TYPE (not a sentinel) - it carries
// structured data about what went wrong, not just a fixed message.
//
// TODO: add fields Container string and Image string, and implement
// Error() string on *InvalidImageError so it satisfies the built-in error
// interface. Format however you like, e.g.:
//
//	fmt.Sprintf("container %s has invalid image %q", e.Container, e.Image)
type InvalidImageError struct {
	Container string
	Image string
}

func (e *InvalidImageError) Error() string {
	return fmt.Sprintf("container %s has invalid image %q", e.Container, e.Image)
}

// SetContainerImage validates newImage is non-empty, then updates the named
// container's Image in place.
//
// TODO: implement.
//   - If the container isn't found, propagate FindContainer's error (wrapped
//     with more context is fine, but must still satisfy errors.Is(err,
//     ErrContainerNotFound)).
//   - If newImage == "", return a wrapped *InvalidImageError (use %w).
//   - Otherwise mutate the container's Image field in place (remember the
//     range-loop-copy gotcha from Lesson 2 - you need index-based mutation,
//     not a copy of the loop variable) and return nil.
func (w *Workload) SetContainerImage(name, newImage string) error {
	if newImage == "" {
		return &InvalidImageError{Container: name, Image: newImage}
	}
	ctrDef, err := w.FindContainer(name)
	if err != nil {
		return err
	}
	for idx, container := range w.Containers {
		if container.Name == ctrDef.Name {
			w.Containers[idx].Image = newImage
		}
	}
	return nil
}

// Validator is an interface with ONE method. Any type - including ones you
// haven't written yet - satisfies it automatically just by having this
// method signature. No `implements` needed.
type Validator interface {
	Validate() error
}

// TODO: implement Validate() on *Workload (pointer receiver) so *Workload
// satisfies Validator. Rule: a Workload is invalid if it has zero
// containers - return a plain error (fmt.Errorf or errors.New) in that case,
// nil otherwise.
func (w *Workload) Validate() error {
	if len(w.Containers) == 0 {
		return errors.New("workload has no containers")
	}
	return nil
}

var _ = errors.New
var _ = fmt.Sprintf

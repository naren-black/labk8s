// Narrated tour, nothing to fill in. Run with:
//
//	go run ./00-go-refresher/03-interfaces-errors/tour
package main

import (
	"errors"
	"fmt"
)

func main() {
	section("1. Structural typing: no 'implements' keyword")
	// Dog and Cat below never mention Animal anywhere in their definition.
	// They satisfy the Animal interface purely because they happen to have
	// a Sound() string method. This is the opposite of Java/C#, where a
	// class must explicitly declare `implements Animal`.
	animals := []Animal{Dog{}, Cat{}}
	for _, a := range animals {
		fmt.Println(a.Sound())
	}

	section("2. error is JUST an interface: `interface{ Error() string }`")
	err := divide(10, 0)
	fmt.Println("err          =", err)
	fmt.Println("err == nil   =", err == nil)
	// Any type with an Error() string method IS an error. That's the whole
	// contract - nothing more.

	section("3. Wrapping errors with %w preserves the chain")
	wrapped := fmt.Errorf("computing report: %w", err)
	fmt.Println("wrapped      =", wrapped)
	// %w (not %v!) keeps a reference to the original err inside wrapped,
	// so tools can still find and inspect the original cause.

	section("4. errors.Is: 'does this chain contain THIS SPECIFIC error value?'")
	fmt.Println("errors.Is(wrapped, ErrDivByZero) =", errors.Is(wrapped, ErrDivByZero))
	fmt.Println("errors.Is(wrapped, ErrNotFound)  =", errors.Is(wrapped, ErrNotFound))
	// ErrDivByZero is a sentinel: a package-level `var Err... = errors.New(...)`
	// that callers compare against by IDENTITY, not by string message.

	section("5. errors.As: 'does this chain contain an error of THIS TYPE?'")
	scaleErr := scaleReplicas(3, -10)
	fmt.Println("scaleErr =", scaleErr)
	var ve *ValidationError
	if errors.As(scaleErr, &ve) {
		fmt.Println("extracted ValidationError, Field =", ve.Field, " Value =", ve.Value)
	}
	// errors.As digs through the wrap chain looking for a value whose
	// CONCRETE TYPE matches *ValidationError, and if found, copies it into ve.

	section("6. An interface you already satisfy without knowing it")
	w := &Workload{Name: "api", Replicas: 3}
	var s Scaler = w // *Workload satisfies Scaler just by having a Scale method
	s.Scale(2)
	fmt.Println("after s.Scale(2) through the interface, w.Replicas =", w.Replicas)
}

func section(title string) {
	fmt.Println()
	fmt.Println("===", title, "===")
}

// --- Section 1 ---

type Animal interface {
	Sound() string
}

type Dog struct{}

func (Dog) Sound() string { return "Woof" }

type Cat struct{}

func (Cat) Sound() string { return "Meow" }

// --- Section 2/3/4 ---

var ErrDivByZero = errors.New("division by zero")
var ErrNotFound = errors.New("not found")

func divide(a, b int) error {
	if b == 0 {
		return ErrDivByZero
	}
	return nil
}

// --- Section 5 ---

// ValidationError is a CUSTOM error type - a struct with an Error() string
// method, so it satisfies the built-in `error` interface. Unlike a sentinel
// (a single value), it carries structured data callers can inspect.
type ValidationError struct {
	Field string
	Value int
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("invalid value for %s: %d", e.Field, e.Value)
}

func scaleReplicas(current, delta int) error {
	result := current + delta
	if result < 0 {
		err := &ValidationError{Field: "replicas", Value: result}
		return fmt.Errorf("scaleReplicas: %w", err)
	}
	return nil
}

// --- Section 6 ---

type Scaler interface {
	Scale(delta int)
}

type Workload struct {
	Name     string
	Replicas int
}

func (w *Workload) Scale(delta int) {
	w.Replicas += delta
	if w.Replicas < 0 {
		w.Replicas = 0
	}
}

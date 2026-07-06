// This file is a NARRATED TOUR, not an exercise. Nothing to fill in —
// just run it and read the output alongside the source, one section at a
// time. Run it with:
//
//	go run ./00-go-refresher/02-pointers-and-methods/tour
package main

import "fmt"

func main() {
	section("1. A plain variable and its address")
	x := 10
	fmt.Println("x        =", x)
	fmt.Println("&x       =", &x, "  <- the ADDRESS where x lives in memory")
	p := &x // p is a *int: "a pointer to an int"
	fmt.Println("p        =", p, "  <- same address, now stored in a pointer variable")
	fmt.Println("*p       =", *p, "  <- dereference: follow the pointer to read the value")
	*p = 99 // mutate through the pointer
	fmt.Println("*p = 99  ->  x =", x, "  <- x changed! p and x refer to the same memory")

	section("2. Pass by value vs pass by pointer")
	n := 5
	addOneByValue(n)
	fmt.Println("after addOneByValue(n):        n =", n, "  <- unchanged, function got a COPY of n")
	addOneByPointer(&n)
	fmt.Println("after addOneByPointer(&n):     n =", n, "  <- changed, function had the real address")

	section("3. Structs are copied by value too")
	c1 := ContainerDefinition{Name: "app", Image: "nginx:1.25"}
	c2 := c1 // this is a full COPY of c1, not a reference to it
	c2.Image = "nginx:1.26"
	fmt.Println("c1.Image =", c1.Image, "  <- untouched")
	fmt.Println("c2.Image =", c2.Image, "  <- only the copy changed")

	section("4. A pointer to a struct shares the SAME struct")
	c3 := &c1 // c3 points AT c1 — no copy this time
	c3.Image = "nginx:1.27"
	fmt.Println("c1.Image =", c1.Image, "  <- changed! c3 and c1 are the same memory")

	section("5. Method with a VALUE receiver: reads, never mutates the original")
	w := NewWorkload("api", "prod", 3)
	fmt.Println(w.String())

	section("6. Method with a POINTER receiver: mutates the original")
	w.Scale(2)
	fmt.Println("after w.Scale(2):", w.String())

	section("7. A slice field: Containers []ContainerDefinition")
	w.AddContainer(ContainerDefinition{Name: "app", Image: "myapp:v1"})
	w.AddContainer(ContainerDefinition{Name: "sidecar", Image: "envoy:v2"})
	fmt.Printf("Containers: %+v\n", w.Containers)

	section("8. THE CLASSIC GOTCHA: the range-loop variable is a COPY")
	for _, ctr := range w.Containers {
		ctr.Image = "OVERWRITTEN" // this only mutates the loop's local copy!
	}
	fmt.Printf("after buggy range-mutate: %+v\n", w.Containers)
	fmt.Println("   ^ unchanged! `ctr` in `for _, ctr := range ...` is a fresh copy each iteration.")

	section("9. The fix: index into the slice directly")
	for i := range w.Containers {
		w.Containers[i].Image = w.Containers[i].Image + "-patched"
	}
	fmt.Printf("after index-based mutate: %+v\n", w.Containers)
	fmt.Println("   ^ changed! w.Containers[i] refers to the real element, not a copy.")
}

func section(title string) {
	fmt.Println()
	fmt.Println("===", title, "===")
}

func addOneByValue(n int) {
	n = n + 1
}

func addOneByPointer(n *int) {
	*n = *n + 1
}

// ContainerDefinition is a stand-in for a real k8s corev1.Container: a
// pod's spec holds a SLICE of these (Pod.Spec.Containers []Container).
type ContainerDefinition struct {
	Name  string
	Image string
}

type Workload struct {
	Name       string
	Namespace  string
	Labels     map[string]string
	Replicas   int
	Containers []ContainerDefinition
}

func NewWorkload(name, namespace string, replicas int) *Workload {
	return &Workload{
		Name:      name,
		Namespace: namespace,
		Replicas:  replicas,
		Labels:    make(map[string]string),
		Containers: make([]ContainerDefinition, 0),
	}
}

// String has a VALUE receiver: it only reads fields, so it doesn't need a pointer.
func (w Workload) String() string {
	return fmt.Sprintf("%s/%s (replicas=%d, containers=%d)", w.Namespace, w.Name, w.Replicas, len(w.Containers))
}

// Scale has a POINTER receiver: it mutates w.Replicas, so the caller needs
// the change to stick.
func (w *Workload) Scale(delta int) {
	w.Replicas += delta
	if w.Replicas < 0 {
		w.Replicas = 0
	}
}

// AddContainer has a POINTER receiver too: append() returns a (possibly new)
// slice header, and we need to write that back into w.Containers.
func (w *Workload) AddContainer(c ContainerDefinition) {
	w.Containers = append(w.Containers, c)
}

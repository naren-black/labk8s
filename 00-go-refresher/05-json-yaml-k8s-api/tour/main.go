// Narrated tour, nothing to fill in. Run with:
//
//	go run ./00-go-refresher/05-json-yaml-k8s-api/tour
package main

import (
	"encoding/json"
	"fmt"

	"sigs.k8s.io/yaml"
)

func main() {
	section("1. Embedding TypeMeta/ObjectMeta: the shape of every k8s object")
	replicas := int32(3)
	pod := Pod{
		TypeMeta:   TypeMeta{Kind: "Pod", APIVersion: "v1"},
		ObjectMeta: ObjectMeta{Name: "web-1", Namespace: "default", Labels: map[string]string{"app": "web"}},
		Spec:       PodSpec{Image: "nginx:1.27", Replicas: &replicas},
	}
	b, _ := json.MarshalIndent(pod, "", "  ")
	fmt.Println(string(b))
	// Notice "kind" and "apiVersion" appear at the TOP level, not nested under
	// a "TypeMeta" key. That's because TypeMeta is EMBEDDED (anonymous field)
	// with tag `json:",inline"`. The word "inline" is actually meaningless to
	// encoding/json - it only recognizes "omitempty" and "string" as options.
	// What ACTUALLY causes the flattening is the EMPTY name before the comma.
	// An embedded struct with an empty-or-absent json name gets its fields
	// promoted to the parent level automatically. ",inline" is a convention
	// borrowed from YAML tooling that people write out of habit - it's
	// harmless here, but it's not doing what it looks like it's doing.

	section("2. Pointers for optional fields: nil vs explicit zero")
	specWithZero := PodSpec{Image: "nginx", Replicas: int32Ptr(0)}
	specWithNil := PodSpec{Image: "nginx", Replicas: nil}
	b1, _ := json.Marshal(specWithZero)
	b2, _ := json.Marshal(specWithNil)
	fmt.Println("Replicas explicitly 0:", string(b1))
	fmt.Println("Replicas unset (nil): ", string(b2))
	// If Replicas were a plain `int` with `omitempty`, BOTH cases above would
	// serialize identically (the field would vanish either way), because
	// omitempty treats the zero value as "absent" - there'd be no way to
	// distinguish "scale to 0 replicas" from "caller didn't set replicas at
	// all". This is exactly why real k8s API types use *int32 for Replicas:
	// a nil pointer means "not set, use the default", and a pointer to 0
	// means "set it to exactly zero". You'll see this pattern constantly in
	// client-go and CRD types.

	section("3. Unmarshaling: field matching is tag-first, then case-insensitive name")
	raw := []byte(`{"kind":"Pod","apiVersion":"v1","metadata":{"name":"from-json"},"spec":{"image":"redis"}}`)
	var decoded Pod
	if err := json.Unmarshal(raw, &decoded); err != nil {
		fmt.Println("unmarshal error:", err)
	}
	fmt.Printf("decoded.ObjectMeta.Name = %q, decoded.Spec.Image = %q\n", decoded.ObjectMeta.Name, decoded.Spec.Image)

	section("4. YAML round-trip: sigs.k8s.io/yaml reuses your JSON tags")
	y, _ := yaml.Marshal(pod)
	fmt.Println(string(y))
	// sigs.k8s.io/yaml works by marshaling to JSON internally, then
	// converting that JSON to YAML - so it honors your `json:"..."` tags
	// with ZERO extra `yaml:"..."` tags needed. This is deliberate: real k8s
	// api types have only json tags, yet kubectl reads/writes them as YAML
	// all day. Contrast with the popular gopkg.in/yaml.v3 package, which
	// knows NOTHING about json tags - it needs its own `yaml:"..."` tags on
	// every field, or it falls back to lowercased Go field names. Mixing the
	// two up is a common source of "why did my YAML field come out wrong"
	// bugs; apimachinery deliberately depends on sigs.k8s.io/yaml to avoid
	// that split entirely.
	var fromYAML Pod
	if err := yaml.Unmarshal(y, &fromYAML); err != nil {
		fmt.Println("yaml unmarshal error:", err)
	}
	fmt.Println("round-tripped Name:", fromYAML.ObjectMeta.Name)

	section("5. Unstructured: map[string]interface{} when you don't have a struct")
	manifest := []byte(`{"apiVersion":"apps/v1","kind":"Deployment","metadata":{"name":"api","labels":{"tier":"backend"}},"spec":{"replicas":5}}`)
	var obj map[string]interface{}
	if err := json.Unmarshal(manifest, &obj); err != nil {
		fmt.Println("unmarshal error:", err)
	}
	// This is exactly the shape client-go's dynamic client and
	// unstructured.Unstructured work with - useful when you're writing
	// generic tooling that handles CRDs you don't have Go types for.
	kind, _ := obj["kind"].(string)
	metadata, _ := obj["metadata"].(map[string]interface{})
	name, _ := metadata["name"].(string)
	fmt.Printf("kind=%q name=%q\n", kind, name)
	// Every level needs its own type assertion with the comma-ok form -
	// there's no compile-time safety here. A typo like obj["Kind"] (wrong
	// case) silently gives you a zero value (empty string, false ok),
	// never a compile error. This is the tradeoff for not needing a
	// pre-defined struct: total flexibility, zero type safety.

	section("6. json.RawMessage: defer parsing part of a document")
	event := []byte(`{"kind":"Pod","spec":{"image":"nginx","replicas":2}}`)
	var envelope struct {
		Kind string          `json:"kind"`
		Spec json.RawMessage `json:"spec"`
	}
	if err := json.Unmarshal(event, &envelope); err != nil {
		fmt.Println("unmarshal error:", err)
	}
	fmt.Println("envelope.Kind =", envelope.Kind)
	fmt.Println("envelope.Spec (still raw bytes) =", string(envelope.Spec))
	// Now that we know Kind == "Pod", we can decode Spec into the RIGHT
	// concrete type. This "peek at the envelope, then decode the payload
	// based on what you saw" pattern is how Kubernetes watch events and
	// admission webhooks work: the outer object tells you what Kind you're
	// looking at, and only then do you know which Go struct to decode the
	// rest into.
	var podSpec PodSpec
	if err := json.Unmarshal(envelope.Spec, &podSpec); err != nil {
		fmt.Println("spec unmarshal error:", err)
	}
	fmt.Printf("decoded spec: %+v\n", podSpec)
}

func section(title string) {
	fmt.Println()
	fmt.Println("===", title, "===")
}

func int32Ptr(i int32) *int32 { return &i }

// --- Types mirroring real k8s API conventions ---

type TypeMeta struct {
	Kind       string `json:"kind,omitempty"`
	APIVersion string `json:"apiVersion,omitempty"`
}

type ObjectMeta struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
}

type PodSpec struct {
	Image    string `json:"image"`
	Replicas *int32 `json:"replicas,omitempty"`
}

type Pod struct {
	TypeMeta   `json:",inline"`
	ObjectMeta ObjectMeta `json:"metadata"`
	Spec       PodSpec    `json:"spec"`
}

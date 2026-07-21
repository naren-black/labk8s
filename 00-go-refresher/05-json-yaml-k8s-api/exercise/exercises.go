// Package k8sapi is Lesson 5's exercise: JSON/YAML marshaling and the shape
// of the Kubernetes API. Read ../tour/main.go first (just run it).
package k8sapi

import (
	"encoding/json"
	"fmt"

	"sigs.k8s.io/yaml"
)

// --- Types mirroring real k8s API conventions (see the tour). ---

type TypeMeta struct {
	Kind       string `json:"kind,omitempty"`
	APIVersion string `json:"apiVersion,omitempty"`
}

type ObjectMeta struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// DeploymentSpec uses a *int32 for Replicas - same nil-vs-explicit-zero
// gotcha covered in the tour.
type DeploymentSpec struct {
	Image    string `json:"image"`
	Replicas *int32 `json:"replicas,omitempty"`
}

type Deployment struct {
	TypeMeta   `json:",inline"`
	ObjectMeta ObjectMeta `json:"metadata"`
	Spec       DeploymentSpec `json:"spec"`
}

// ServiceSpec is a second, unrelated spec type, used by DecodeByKind below
// to simulate handling more than one Kind polymorphically.
type ServiceSpec struct {
	Port int `json:"port"`
}

// ParseManifest unmarshals a JSON manifest (like the ones kubectl sends)
// into a *Deployment.
//
// TODO: implement with json.Unmarshal. Return the parse error unchanged if
// unmarshaling fails.
func ParseManifest(data []byte) (*Deployment, error) {
	panic("TODO: implement ParseManifest")
}

// RenderYAML converts a *Deployment into YAML bytes.
//
// TODO: implement using yaml.Marshal from sigs.k8s.io/yaml (already
// imported). Note this package reuses your `json:"..."` struct tags - you
// don't need any yaml-specific tags on the types above.
func RenderYAML(d *Deployment) ([]byte, error) {
	panic("TODO: implement RenderYAML")
}

// EffectiveReplicas returns the deployment's configured replica count, or
// defaultReplicas if Replicas was never set.
//
// TODO: implement.
//   - If d.Spec.Replicas is nil, return defaultReplicas.
//   - Otherwise return *d.Spec.Replicas - INCLUDING when it's explicitly 0.
//     (0 is a valid, deliberate choice - "scale to zero" - and must NOT be
//     treated the same as "unset". This is the whole reason Replicas is a
//     pointer instead of a plain int32.)
func EffectiveReplicas(d *Deployment, defaultReplicas int32) int32 {
	panic("TODO: implement EffectiveReplicas")
}

// ExtractLabel digs a single label value out of a raw JSON manifest without
// using the Deployment struct at all - the "unstructured" pattern from
// section 5 of the tour.
//
// TODO: implement.
//   - Unmarshal manifest into a map[string]interface{}.
//   - Navigate down to obj["metadata"] (map[string]interface{}), then to
//     its "labels" key (map[string]interface{}), then to labels[key].
//   - Use the comma-ok form of every type assertion. If any step is
//     missing or the wrong type (not a map, key absent, value not a
//     string), return ("", false) - do NOT panic.
func ExtractLabel(manifest []byte, key string) (string, bool) {
	panic("TODO: implement ExtractLabel")
}

// DecodeByKind peeks at the "kind" field of a raw JSON document, then
// decodes the "spec" field into the concrete type that matches - the
// envelope pattern from section 6 of the tour (used constantly by watch
// events and admission webhooks, which see many different Kinds over the
// same wire format).
//
// TODO: implement.
//   - Unmarshal data into an anonymous struct with a Kind string field
//     (tag `json:"kind"`) and a Spec json.RawMessage field (tag
//     `json:"spec"`).
//   - If Kind == "Deployment", unmarshal Spec into a DeploymentSpec and
//     return a *DeploymentSpec.
//   - If Kind == "Service", unmarshal Spec into a ServiceSpec and return a
//     *ServiceSpec.
//   - Otherwise return nil and an error like
//     fmt.Errorf("unknown kind %q", kind).
func DecodeByKind(data []byte) (interface{}, error) {
	panic("TODO: implement DecodeByKind")
}

var _ = json.Unmarshal
var _ = yaml.Marshal
var _ = fmt.Errorf

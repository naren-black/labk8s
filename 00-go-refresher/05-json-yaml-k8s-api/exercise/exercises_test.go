package k8sapi

import (
	"strings"
	"testing"
)

func TestParseManifest(t *testing.T) {
	data := []byte(`{
		"kind": "Deployment",
		"apiVersion": "apps/v1",
		"metadata": {"name": "api", "namespace": "prod", "labels": {"tier": "backend"}},
		"spec": {"image": "myapp:v1", "replicas": 3}
	}`)

	d, err := ParseManifest(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Kind != "Deployment" || d.APIVersion != "apps/v1" {
		t.Fatalf("got Kind=%q APIVersion=%q, want Deployment/apps/v1", d.Kind, d.APIVersion)
	}
	if d.ObjectMeta.Name != "api" || d.ObjectMeta.Namespace != "prod" {
		t.Fatalf("got Name=%q Namespace=%q, want api/prod", d.ObjectMeta.Name, d.ObjectMeta.Namespace)
	}
	if d.ObjectMeta.Labels["tier"] != "backend" {
		t.Fatalf("got Labels[tier]=%q, want %q", d.ObjectMeta.Labels["tier"], "backend")
	}
	if d.Spec.Image != "myapp:v1" {
		t.Fatalf("got Image=%q, want %q", d.Spec.Image, "myapp:v1")
	}
	if d.Spec.Replicas == nil || *d.Spec.Replicas != 3 {
		t.Fatalf("got Replicas=%v, want pointer to 3", d.Spec.Replicas)
	}
}

func TestRenderYAML(t *testing.T) {
	replicas := int32(2)
	d := &Deployment{
		TypeMeta:   TypeMeta{Kind: "Deployment", APIVersion: "apps/v1"},
		ObjectMeta: ObjectMeta{Name: "api", Namespace: "prod"},
		Spec:       DeploymentSpec{Image: "myapp:v1", Replicas: &replicas},
	}

	y, err := RenderYAML(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := string(y)
	for _, want := range []string{"kind: Deployment", "name: api", "image: myapp:v1", "replicas: 2"} {
		if !strings.Contains(out, want) {
			t.Fatalf("rendered YAML missing %q, got:\n%s", want, out)
		}
	}
}

func TestEffectiveReplicas(t *testing.T) {
	zero := int32(0)
	five := int32(5)

	cases := []struct {
		name     string
		replicas *int32
		def      int32
		want     int32
	}{
		{"unset uses default", nil, 3, 3},
		{"explicit zero is NOT the default", &zero, 3, 0},
		{"explicit value wins", &five, 3, 5},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d := &Deployment{Spec: DeploymentSpec{Replicas: c.replicas}}
			got := EffectiveReplicas(d, c.def)
			if got != c.want {
				t.Fatalf("got %d, want %d", got, c.want)
			}
		})
	}
}

func TestExtractLabel(t *testing.T) {
	manifest := []byte(`{"metadata":{"name":"api","labels":{"tier":"backend"}}}`)

	v, ok := ExtractLabel(manifest, "tier")
	if !ok || v != "backend" {
		t.Fatalf("got (%q, %v), want (%q, true)", v, ok, "backend")
	}

	_, ok = ExtractLabel(manifest, "does-not-exist")
	if ok {
		t.Fatalf("expected ok=false for a missing label key")
	}

	noLabels := []byte(`{"metadata":{"name":"api"}}`)
	_, ok = ExtractLabel(noLabels, "tier")
	if ok {
		t.Fatalf("expected ok=false when metadata has no labels map at all")
	}

	noMetadata := []byte(`{}`)
	_, ok = ExtractLabel(noMetadata, "tier")
	if ok {
		t.Fatalf("expected ok=false when metadata is missing entirely")
	}
}

func TestDecodeByKind(t *testing.T) {
	depData := []byte(`{"kind":"Deployment","spec":{"image":"nginx","replicas":4}}`)
	got, err := DecodeByKind(depData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	dep, ok := got.(*DeploymentSpec)
	if !ok {
		t.Fatalf("got %T, want *DeploymentSpec", got)
	}
	if dep.Image != "nginx" || dep.Replicas == nil || *dep.Replicas != 4 {
		t.Fatalf("got %+v, want Image=nginx Replicas=4", dep)
	}

	svcData := []byte(`{"kind":"Service","spec":{"port":8080}}`)
	got, err = DecodeByKind(svcData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	svc, ok := got.(*ServiceSpec)
	if !ok {
		t.Fatalf("got %T, want *ServiceSpec", got)
	}
	if svc.Port != 8080 {
		t.Fatalf("got Port=%d, want 8080", svc.Port)
	}

	_, err = DecodeByKind([]byte(`{"kind":"Widget","spec":{}}`))
	if err == nil {
		t.Fatalf("expected an error for an unknown kind, got nil")
	}
}

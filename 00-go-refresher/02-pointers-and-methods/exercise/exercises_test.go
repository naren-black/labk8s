package pointersmethods

import "testing"

func TestNewWorkload(t *testing.T) {
	w := NewWorkload("api", "prod", map[string]string{"app": "api"})
	if w.Name != "api" || w.Namespace != "prod" {
		t.Fatalf("got Name=%q Namespace=%q, want api/prod", w.Name, w.Namespace)
	}
	if w.Labels["app"] != "api" {
		t.Fatalf("got Labels[app]=%q, want %q", w.Labels["app"], "api")
	}
	if w.Replicas != 1 {
		t.Fatalf("got Replicas=%d, want 1 (NewWorkload hardcodes this)", w.Replicas)
	}
	if len(w.Containers) != 2 {
		t.Fatalf("got %d default containers, want 2 (ubuntu, nginx)", len(w.Containers))
	}
}

func TestAddContainer(t *testing.T) {
	w := NewWorkload("api", "prod", map[string]string{"app": "api"})
	before := len(w.Containers)

	w.AddContainer(ContainerDefinition{Name: "mysql", Image: "mysql:latest", Ports: []int{3306}})

	if len(w.Containers) != before+1 {
		t.Fatalf("got %d containers, want %d", len(w.Containers), before+1)
	}
	last := w.Containers[len(w.Containers)-1]
	if last.Name != "mysql" || last.Image != "mysql:latest" || len(last.Ports) != 1 || last.Ports[0] != 3306 {
		t.Fatalf("got last container %+v, want Name=mysql Image=mysql:latest Ports=[3306]", last)
	}
}

func TestUpdateContainerImagebyName(t *testing.T) {
	w := NewWorkload("api", "prod", map[string]string{"app": "api"})

	w.UpdateContainerImagebyName("nginx", "nginx:1.27")

	var found bool
	for _, c := range w.Containers {
		if c.Name == "nginx" {
			found = true
			if c.Image != "nginx:1.27" {
				t.Fatalf("got nginx Image=%q, want %q", c.Image, "nginx:1.27")
			}
		}
	}
	if !found {
		t.Fatalf("no container named nginx found in default set - did the default containers change?")
	}

	// Updating a name that doesn't exist should be a silent no-op, not a panic.
	w.UpdateContainerImagebyName("does-not-exist", "whatever")
}

func TestTotalPorts(t *testing.T) {
	w := NewWorkload("api", "prod", map[string]string{"app": "api"})
	// Defaults: ubuntu (1 port) + nginx (1 port) = 2.
	if got := w.TotalPorts(); got != 2 {
		t.Fatalf("got TotalPorts()=%d, want 2 for the default containers", got)
	}

	w.AddContainer(ContainerDefinition{Name: "mysql", Image: "mysql:latest", Ports: []int{3306}})
	if got := w.TotalPorts(); got != 3 {
		t.Fatalf("got TotalPorts()=%d after adding mysql, want 3", got)
	}
}

func TestScaleUpDown(t *testing.T) {
	w := NewWorkload("api", "prod", map[string]string{"app": "api"})
	w.ScaleUp(2)
	if w.Replicas != 3 {
		t.Fatalf("after ScaleUp(2), got Replicas=%d, want 3", w.Replicas)
	}
	w.ScaleDown(1)
	if w.Replicas != 2 {
		t.Fatalf("after ScaleDown(1), got Replicas=%d, want 2", w.Replicas)
	}
}

func TestPatchContainerImageByName(t *testing.T) {
	// NOTE: despite the name, this method currently patches .Name, not .Image.
	// This test documents CURRENT behavior - see the flagged question above
	// about whether that's intentional.
	w := NewWorkload("api", "prod", map[string]string{"app": "api"})
	w.patchContainerImageByName("patched")

	for _, c := range w.Containers {
		if c.Name != "ubuntu-patched" && c.Name != "nginx-patched" {
			t.Fatalf("got container Name=%q, want a name ending in -patched", c.Name)
		}
	}
}

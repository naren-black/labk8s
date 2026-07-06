package pointersmethods

import (
	"fmt"
	"encoding/json"
)

type ContainerDefinition struct {
	Name string
	Image string
	Ports []int
}

type Workload struct {
	Name string
	Namespace string
	Labels     map[string]string
	Replicas   int
	Containers []ContainerDefinition
}

func main() {
	workload := NewWorkload("workload1", "default", map[string]string{"app": "workload1"})
	workload.printJson()
	workload.ScaleUp(1)
	workload.printJson()
	workload.ScaleDown(1)
	workload.printJson()
	workload.AddContainer(getContainerDefinition("mysql", "mysql:latest", []int{3306}))
	workload.printJson()
	workload.patchContainerImageByName("patched4Life")
	workload.printJson()
}

func NewWorkload(name, namespace string, labels map[string]string) *Workload {
	return &Workload{
		Name: name,
		Namespace: namespace,
		Labels: labels,
		Replicas: 1,
		Containers: []ContainerDefinition{getContainerDefinition("ubuntu", "ubuntu:latest", []int{8080}), getContainerDefinition("nginx", "nginx:latest", []int{80})},
	}
}

func (workload *Workload) printJson() {
	b, err := json.MarshalIndent(workload, "", "  ")
	if err != nil {
		fmt.Println("error marshalling workload: ", err)
		return
	}
	fmt.Println("workload value is: ", string(b))
}

func (workload *Workload) patchContainerImageByName(patchName string) {
	
	for idx, cnt := range workload.Containers {
		fmt.Printf("Patching container image %s with %s\n", workload.Containers[idx].Name, patchName)
		workload.Containers[idx].Name = fmt.Sprintf("%s-%s", cnt.Name, patchName)
	}
}

func (workload *Workload) UpdateContainerImagebyName(name string, image string) {
	for idx, cnt := range workload.Containers {
		if cnt.Name == name {
			workload.Containers[idx].Image = image
			return
		}
	}
}

func (workload *Workload) TotalPorts() int {
	total := 0
	for _, cnt := range workload.Containers {
		total += len(cnt.Ports)
	}
	return total
}

func (workload *Workload) ScaleUp(num int) {
	fmt.Println("Scaling up workload by ", num)
	workload.Replicas = workload.Replicas + num
}

func (workload *Workload) ScaleDown(num int) {
	fmt.Println("Scaling down workload by ", num)
	workload.Replicas = workload.Replicas - num
}

func (workload *Workload) AddContainer(container ContainerDefinition) {
	fmt.Println("Adding container ", container.Name, " to workload")
	workload.Containers = append(workload.Containers, container)
}

func getContainerDefinition(name, image string, ports []int) ContainerDefinition {
	return ContainerDefinition{
		Name: name,
		Image: image,
		Ports: ports,
	}
}

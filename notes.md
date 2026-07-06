What we're installing, and why

On all 3 nodes:

┌────────────┬────────────────────────┬─────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│ Component  │       What it is       │                                           Why every node needs it                                           │
├────────────┼────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ containerd │ The container runtime  │ Kubernetes doesn't run containers itself — kubelet talks to a CRI-compliant runtime over gRPC to pull       │
│            │ (CRI)                  │ images and start/stop containers. containerd is the standard choice (Docker uses it internally too).        │
├────────────┼────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│            │ Per-node agent, runs   │ Watches the API server for Pods assigned to its node, tells containerd to run them, reports status back.    │
│ kubelet    │ as a systemd service   │ This is the only k8s component that runs as a native OS process rather than a container — it has to, since  │
│            │                        │ it's the thing responsible for starting containers in the first place.                                      │
├────────────┼────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ kubeadm    │ A one-shot             │ You run it once (init on the control plane, join on workers). It generates certs, kubeconfigs, and static   │
│            │ bootstrapping CLI      │ pod manifests for you. It is not a long-running service — it does its job and exits.                        │
└────────────┴────────────────────────┴─────────────────────────────────────────────────────────────────────────────────────────────────────────────┘

On the control-plane node only, kubeadm init will additionally start these as static pods (containers that kubelet runs directly from manifest files on disk, not scheduled via the API server — this is how the control plane bootstraps itself before an API server even exists to schedule anything):

┌─────────────────────────┬──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│        Component        │                                                           Role                                                           │
├─────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ etcd                    │ The cluster's entire state lives here — every Pod, Service, ConfigMap, everything. A key-value store, nothing more.      │
├─────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ kube-apiserver          │ The front door. Every other component — kubectl, kubelet, scheduler, controller-manager — talks only to the API server,  │
│                         │ never directly to etcd or each other.                                                                                    │
├─────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ kube-controller-manager │ Runs the reconciliation loops (node health, ReplicaSet scaling, etc.) — the built-in "controllers" that watch desired    │
│                         │ vs. actual state and correct drift. This is the direct ancestor of the custom controllers we'll write by hand later.     │
├─────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ kube-scheduler          │ Watches for unscheduled Pods and assigns them to a node based on resources/constraints.                                  │
└─────────────────────────┴──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘

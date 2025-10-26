# Daggerk3 - K3s ArgoCD Integration Test

A Dagger module that spins up a K3s Kubernetes cluster and deploys ArgoCD as an integration test using the [k3s daggerverse module](https://daggerverse.dev/mod/github.com/marcosnils/daggerverse/k3s).

## ✅ Status: Working!

This module successfully:
- Creates K3s clusters in Dagger containers
- Deploys ArgoCD to the cluster
- Waits for all deployments to be ready
- Returns deployment status

## Features

- **Full K3s Cluster**: Uses the k3s daggerverse module with proper cgroup handling
- **ArgoCD Deployment**: Deploys ArgoCD v2.13.2 with all components
- **Health Checks**: Waits for deployments to be available before completing
- **Multiple Functions**: IntegrationTest, DeployArgoCD, GetArgoCDStatus

## Prerequisites

- [Dagger](https://docs.dagger.io/install) v0.18.16 or later
- Docker running locally

## Installation

```bash
git clone <repo-url>
cd daggerk3
```

## Usage

### Run Complete Integration Test

Runs the full test: creates cluster, deploys ArgoCD, waits for ready, returns status.

```bash
dagger call integration-test
```

With a custom cluster name:

```bash
dagger call integration-test --cluster-name my-test
```

**Expected output:**
```
Integration Test Successful!

Cluster: argocd-test
Timestamp: 2025-10-26T13:55:30Z

=== Deployment Output ===
customresourcedefinition.apiextensions.k8s.io/applications.argoproj.io created
serviceaccount/argocd-application-controller created
...

=== Wait Output ===
deployment.apps/argocd-applicationset-controller condition met
deployment.apps/argocd-dex-server condition met
...

=== ArgoCD Pods ===
NAME                                                READY   STATUS    RESTARTS   AGE
argocd-application-controller-0                     1/1     Running   0          4m1s
argocd-applicationset-controller-794d89b65b-mcz9c   1/1     Running   0          4m1s
...
```

### Deploy ArgoCD Only

Just deploys ArgoCD without waiting:

```bash
dagger call deploy-argo-cd --cluster-name test-cluster
```

### Check ArgoCD Status

Check the status of ArgoCD pods in a cluster:

```bash
dagger call get-argo-cdstatus --cluster-name test-cluster
```

## How It Works

1. **K3s Module**: Uses `dag.K3S(name)` from the daggerverse to create a cluster
2. **Server Start**: Explicitly starts the server with `server.Start(ctx)`
3. **Kubectl Commands**: Uses the k3s module's `Kubectl()` method for all operations
4. **Cgroup Handling**: The k3s module handles cgroup v2 setup automatically
5. **Service Lifecycle**: Dagger manages the k3s service lifecycle

## Available Functions

### integration-test

Runs a complete end-to-end test:
- Creates K3s cluster
- Deploys ArgoCD
- Waits for deployments to be ready (5min timeout)
- Returns full status report

**Parameters:**
- `--cluster-name` (optional, default: "argocd-test")

### deploy-argo-cd

Deploys ArgoCD to a cluster without waiting for ready status.

**Parameters:**
- `--cluster-name` (optional, default: "test-cluster")

### get-argo-cdstatus

Gets the current status of ArgoCD pods.

**Parameters:**
- `--cluster-name` (optional, default: "test-cluster")

## Project Structure

```
.
├── dagger/
│   ├── main.go              # Main Dagger module implementation
│   ├── dagger.gen.go        # Generated Dagger code
│   ├── go.mod               # Module dependencies
│   └── internal/            # Generated internal types
├── dagger.json              # Dagger configuration with k3s dependency
├── go.mod                   # Root Go module
├── IMPLEMENTATION_PLAN.md   # Development journey and lessons learned
├── LICENSE                  # Apache-2.0 license
└── README.md                # This file
```

## Implementation Details

### Key Code Pattern

```go
// Create K3s cluster
k3s := dag.K3S(clusterName)

// Start the server explicitly
server := k3s.Server()
_, err := server.Start(ctx)

// Now use kubectl
_, err = k3s.Kubectl("create namespace argocd").Sync(ctx)
deployOutput, err := k3s.Kubectl("apply -f manifests.yaml").Stdout(ctx)
```

### Why It Works

The k3s daggerverse module:
1. **Handles cgroup v2** with a custom entrypoint script that evacuates the root cgroup
2. **Manages kubeconfig** automatically in a cache volume
3. **Provides kubectl wrapper** that handles service binding internally
4. **Caches cluster state** for faster subsequent operations

## Configuration

- **ArgoCD Version**: v2.13.2 (configurable in `dagger/main.go`)
- **Namespace**: argocd (configurable in `dagger/main.go`)
- **Wait Timeout**: 300s for deployments to become ready

## Dependencies

This module uses:
- [k3s Dagger module](https://github.com/marcosnils/daggerverse/tree/main/k3s) - v0.1.10
  - Handles K3s cluster creation and management
  - Provides kubectl command execution
  - Manages kubeconfig and cluster state

## Performance

- **First run**: ~4-5 minutes (downloads images, starts cluster, deploys ArgoCD)
- **Subsequent runs**: ~1-2 minutes (leverages Dagger caching)
- **Cluster startup**: ~10-15 seconds
- **ArgoCD deployment**: ~3-4 minutes (including wait for ready)

## Troubleshooting

### Timeout waiting for deployments

If deployments don't become ready within 5 minutes, the test will fail. This can happen if:
- Docker resources are constrained
- Network is slow downloading images
- Previous cluster state is corrupted

**Solution**: Increase timeout in `main.go` or clean Dagger cache:
```bash
dagger core version  # Shows cache location
rm -rf ~/.cache/dagger  # Nuclear option
```

### "k3s.yaml not ready" messages

This means the k3s server hasn't finished initializing. The module will wait automatically. If it persists beyond 30 seconds, there may be an issue with Docker or system resources.

## Development

List available functions:

```bash
dagger functions
```

View function details:

```bash
dagger call integration-test --help
```

Regenerate Dagger code after changes:

```bash
dagger develop
```

## References

- [Dagger Documentation](https://docs.dagger.io/)
- [K3s Documentation](https://docs.k3s.io/)
- [K3s Daggerverse Module](https://daggerverse.dev/mod/github.com/marcosnils/daggerverse/k3s)
- [ArgoCD Documentation](https://argo-cd.readthedocs.io/)
- [IMPLEMENTATION_PLAN.md](./IMPLEMENTATION_PLAN.md) - Full development journey

## License

Apache-2.0

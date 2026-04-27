# Development

## Prerequisites
- [Go](https://go.dev/) (1.26+)
- [Helm](https://helm.sh/)
- [Skaffold](https://skaffold.dev/)
- A Kubernetes cluster (e.g., [kind](https://kind.sigs.k8s.io/))

Development tools like `ko`, `golangci-lint`, and `setup-envtest` are managed automatically via the Go toolchain (`go tool`).

## Project Structure
- `cmd/stackit-pod-identity-webhook`: Entry point for the webhook manager.
- `pkg/webhook`: Core mutation logic and admission handler.
- `charts/stackit-pod-identity-webhook`: Helm chart for deployment.
- `test/integration`: Integration tests using `envtest` and Ginkgo.
- `hack`: Scripts and tooling configuration.

## Makefile Targets
To see all available targets and their descriptions, run `make help`.

# Local Deployment (kind)
The easiest way to deploy locally is using the provided Makefile targets, which automatically manage the cluster and install `cert-manager` and the webhook:

```bash
make kind-up
make skaffold-dev
```
## Testing Mutation

After deploying the webhook, you can verify it's working by applying a simple test pod:

1. Apply the test resources:
   ```bash
   kubectl apply -f examples/mutation-test.yaml
   ```
2. Verify the mutation on the Pod:
   ```bash
   kubectl get pod mutation-test-pod -o yaml
   ```
   You should see injected environment variables (like `STACKIT_TOKEN_PATH`) and volume mounts.

For a more comprehensive test that includes authentication with the STACKIT SDK, see the [Example Workload Documentation](example-workload.md).


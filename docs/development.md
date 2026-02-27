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
- `make verify`: Runs formatting checks, module tidying, linting, and all tests (default target).
- `make check`: Runs linting and tests.
- `make test`: Runs unit and integration tests using Ginkgo.
- `make build`: Builds the manager binary in `bin/manager`.
- `make image`: Builds the container image using `ko`.
- `make skaffold-dev`: Runs `skaffold dev` with `cert-manager` enabled for local development.

# Local Deployment (kind)
The easiest way to deploy locally is using the provided Makefile target, which automatically installs `cert-manager` and the webhook:

```bash
kind create cluster
make skaffold-dev
```
## Testing Mutation

After deploying the webhook you can check if it works:

1. Apply the sample resources:
   ```bash
   kubectl apply -f samples/test-identity.yaml
   ```
2. Verify the mutation on the Pod:
   ```bash
   kubectl get pod test-pod -o yaml
   ```


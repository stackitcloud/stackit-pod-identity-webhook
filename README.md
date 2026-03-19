# stackit-pod-identity-webhook

A Kubernetes mutating admission webhook that injects STACKIT workload identity compatible tokens into pods.
This enables "secret-less" authentication for workloads running in Kubernetes clusters which need to authenticate
for the STACKIT API.

## Use with SKE (STACKIT Kubernetes Engine)

If you're using SKE you don't need to setup this webhook in your cluster. It comes preconfigured with every SKE cluster.
Just set the `workload-identity.stackit.cloud/service-account-email` for your desired `ServiceAccount` ([see documentation](https://docs.stackit.cloud/platform/access-and-identity/service-accounts/how-tos/manage-service-accounts/)) and the
projected volume and environment variables for the SDK will be injected into your pod spec using the annotated `ServiceAccount`.

Make sure to configure the [service account federation](https://docs.stackit.cloud/platform/access-and-identity/service-accounts/how-tos/manage-service-account-federations)
for the service account you want to use from within your cluster. Therefore you need to specify the issuer url which is dependant on the environment
`https://discovery.${BASE_DOMAIN}/projects/ondemand/shoots/<shoot-uid>/issuer` (also see `Shoot.status.advertisedAddresses`).

The `workload-identity.stackit.cloud/audience` and `workload-identity.stackit.cloud/service-account-email` annotations of the `ServiceAccount` in your cluster must match the configuration in the portal.
Grant the service account the permissions necessary for your use-case, e.g. `reader`.

## Features

- **[Projected ServiceAccount Tokens](https://kubernetes.io/docs/concepts/storage/projected-volumes/#serviceaccounttoken):** Injects projected volumes with audience-bound tokens.
- **SDK Configuration:** Automatically sets environment variables (`STACKIT_SERVICE_ACCOUNT_EMAIL`, etc.) for the [STACKIT SDK](https://github.com/stackitcloud/stackit-sdk-go).
- **Configurable:** Supports per-ServiceAccount configuration via annotations for audience, token expiration, and IDP endpoints.

## Supported Annotations

### ServiceAccount Annotations

The webhook looks for the following annotations on `ServiceAccounts`:

| Annotation                                                                 | Description                                                                                                                                                                                             | Default                      |
| -------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------- |
| `workload-identity.stackit.cloud/service-account-email`                    | Enables workload identity and specifies the STACKIT Service Account email.                                                                                                                              | Required to enable           |
| `workload-identity.stackit.cloud/audience`                                 | Audience for the projected token. The audience is part of the configured trust relationship at the Identity Provider. For use with the STACKIT Identity Provider you usually don't need to change this. | `sts.accounts.stackit.cloud` |
| `workload-identity.stackit.cloud/service-account-token-expiration-seconds` | Validity of the projected Kubernetes Service Account token.                                                                                                                                             | `600`                        |
| `workload-identity.stackit.cloud/idp-token-expiration-seconds`             | Validity of the token issued by the Identity Provider.                                                                                                                                                  | SDK default                  |
| `workload-identity.stackit.cloud/idp-token-endpoint`                       | The Identity Provider endpoint for token exchange.                                                                                                                                                      | SDK default                  |
| `workload-identity.stackit.cloud/federated-token-file`                     | Path to the projected token file. Also changes mount path.                                                                                                                                              | SDK default                  |

## Exclusion Logic

The webhook can be configured to skip mutation for specific resources using labels:

- **Skip a specific Pod:** Add the label `workload-identity.stackit.cloud/skip-pod-identity-webhook: "true"` to the Pod.
- **Skip a whole Namespace:** Add the label `workload-identity.stackit.cloud/skip-pod-identity-webhook: "true"` to the Namespace. All Pods in this namespace will be ignored.

> [!NOTE]
> Label-based filtering and the exclusion of the `kube-system` namespace depend on the `MutatingWebhookConfiguration` selectors. These are pre-configured in the provided Helm chart and come as a standard feature in STACKIT Kubernetes Engine (SKE).

# Deployment

## Production Deployment

For production, it is recommended to use [cert-manager](https://cert-manager.io/) to manage the webhook's TLS certificates.

1. **Install cert-manager** in your cluster.
2. **Configure an Issuer or ClusterIssuer.**
3. **Deploy the Helm chart** with cert-manager enabled:

   ```bash
   helm install stackit-pod-identity-webhook ./charts/stackit-pod-identity-webhook 
     --set certmanager.enabled=true 
     --set certmanager.issuerName=<your-issuer-name> 
     --set certmanager.issuerKind=<Issuer-or-ClusterIssuer>
   ```

When `certmanager.enabled` is `true`, the chart will:

- Create a `Certificate` resource.
- Use `cert-manager`'s CA injector to automatically populate the `caBundle` in the `MutatingWebhookConfiguration`.

## Development

For information on how to develop and test the webhook, please see [docs/development.md](docs/development.md).

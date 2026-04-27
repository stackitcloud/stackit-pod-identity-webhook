# Stackit Workload Identity Example App

This directory contains a simple Go application that demonstrates how to use **STACKIT Workload Identity** within a Kubernetes cluster.

## Overview

The application uses the [STACKIT SDK for Go](https://github.com/stackitcloud/stackit-sdk-go) to interact with the STACKIT SKE API. Specifically, it lists the available Kubernetes versions for a given region.

When running in a cluster with the `stackit-pod-identity-webhook` installed, the application automatically uses the identity injected into the Pod via the configured ServiceAccount.

> [!NOTE]  
> SKE Users: Workload identity for SKE clusters is currently in development. It will not be necessary to manually install or manage the webhook.

## Prerequisites

- A STACKIT project.
- A STACKIT Service Account (no permissions or role assignments required)
- The `stackit-pod-identity-webhook` installed in your cluster.
- Rudimentary knowledge about OIDC and identity federation

## Federation Configuration

Refer to the STACKIT IdP documentation
[Create, manage and delete federated identity providers](https://docs.stackit.cloud/platform/access-and-identity/service-accounts/how-tos/manage-service-account-federations/).

Provide your clusters `serviceAccountIssuer` URL. If you're using SKE you can get the URL from your cluster status via 
[API](https://docs.api.eu01.stackit.cloud/documentation/ske#tag/Cluster/operation/SkeService_GetCluster), the STACKIT CLI
or by inspecting your clusters status in the portal.

> [!NOTE]
> We recommended to create an assertion for the `sub` claim to limit the usage of the STACKIT Service Account to a single Kubernetes Service Account.

Create an assertion with the `sub` claim. The value should equal `system:serviceaccount:<namespace>:<k8s-serviceaccount>`.
If [stackit-workload-identity-example-app.yaml](../examples/stackit-workload-identity-example-app.yaml) is deployed to the `default`
namespace the value would be `system:serviceaccount:default:stackit-workload-identity-example-app-sa`.

For more details about the claims withing the cluster's JWT see the Kubernetes docs about 
[Schema for service account private claims](https://kubernetes.io/docs/reference/access-authn-authz/service-accounts-admin/#schema-for-service-account-private-claims).

## Running the Example

1. Update the ServiceAccount annotation in `examples/stackit-workload-identity-example-app.yaml` with your STACKIT Service Account email.
2. Apply the manifest:

```bash
kubectl apply -f examples/stackit-workload-identity-example-app.yaml
```

3. Check the status and logs of the Job:

```bash
kubectl describe job/stackit-workload-identity-example-app
kubectl logs job/stackit-workload-identity-example-app
```

## How it works

The webhook injects the following into the Pod:
- **Environment Variables**: `STACKIT_SERVICE_ACCOUNT_EMAIL` (and others required by the SDK).
- **Volume Mounts**: A projected volume containing the projected ServiceAccount token.

The STACKIT SDK automatically detects these settings and uses them to exchange the Kubernetes
token for a STACKIT token, allowing the application to authenticate against STACKIT APIs without
requiring manual secret management.

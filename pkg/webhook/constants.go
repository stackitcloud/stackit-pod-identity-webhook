package webhook

// Annotations
const (
	// AnnotationServiceAccountEmail is the ServiceAccount annotation that holds the email of the STACKIT service account to impersonate.
	AnnotationServiceAccountEmail = "workload-identity.stackit.cloud/service-account-email"
	// AnnotationAudience is the ServiceAccount annotation that sets the audience of the projected service account token.
	AnnotationAudience = "workload-identity.stackit.cloud/audience"
	// AnnotationServiceAccountTokenExpiration is the ServiceAccount annotation that sets the lifetime of the projected service account token in seconds.
	AnnotationServiceAccountTokenExpiration = "workload-identity.stackit.cloud/service-account-token-expiration-seconds"
	// AnnotationIDPTokenExpiration is the ServiceAccount annotation that sets the lifetime of the token issued by the identity provider in seconds.
	AnnotationIDPTokenExpiration = "workload-identity.stackit.cloud/idp-token-expiration-seconds"
	// AnnotationIDPTokenEndpoint is the ServiceAccount annotation that overrides the identity provider's token endpoint.
	AnnotationIDPTokenEndpoint = "workload-identity.stackit.cloud/idp-token-endpoint"
	// AnnotationFederatedTokenFile is the ServiceAccount annotation that overrides the path of the federated token file inside the container.
	AnnotationFederatedTokenFile = "workload-identity.stackit.cloud/federated-token-file"

	// AnnotationSkipWebhook is the Pod annotation that, when set, makes the webhook skip mutating the Pod.
	AnnotationSkipWebhook = "workload-identity.stackit.cloud/skip-pod-identity-webhook"
)

// Defaults
const (
	// DefaultAudience is the audience used when AnnotationAudience is not set.
	DefaultAudience = "sts.accounts.stackit.cloud"
	// DefaultServiceAccountTokenExpiration is the token lifetime in seconds used when AnnotationServiceAccountTokenExpiration is not set.
	DefaultServiceAccountTokenExpiration = "600"
	// DefaultVolumeName is the name of the projected volume that holds the service account token.
	DefaultVolumeName = "stackit-workload-identity"
	// DefaultMountPath is the directory where the token volume is mounted in the container.
	DefaultMountPath = "/var/run/secrets/stackit.cloud/serviceaccount"
	// DefaultTokenPath is the file name of the token within DefaultMountPath.
	DefaultTokenPath = "token"
)

// Environment variable names
//
//nolint:gosec // G101: environment variable names, not credentials
const (
	// EnvFederatedTokenFile is the environment variable that holds the path of the federated token file.
	EnvFederatedTokenFile = "STACKIT_FEDERATED_TOKEN_FILE"
	// EnvServiceAccountEmail is the environment variable that holds the email of the STACKIT service account.
	EnvServiceAccountEmail = "STACKIT_SERVICE_ACCOUNT_EMAIL"
	// EnvIDPTokenEndpoint is the environment variable that holds the identity provider's token endpoint.
	EnvIDPTokenEndpoint = "STACKIT_IDP_TOKEN_ENDPOINT"
	// EnvIDPTokenExpirationSeconds is the environment variable that holds the lifetime of the identity provider token in seconds.
	EnvIDPTokenExpirationSeconds = "STACKIT_IDP_TOKEN_EXPIRATION_SECONDS"
)

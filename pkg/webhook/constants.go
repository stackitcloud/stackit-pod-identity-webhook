package webhook

const (
	// ServiceAccount annotations
	AnnotationServiceAccountEmail           = "workload-identity.stackit.cloud/service-account-email"
	AnnotationAudience                      = "workload-identity.stackit.cloud/audience"
	AnnotationServiceAccountTokenExpiration = "workload-identity.stackit.cloud/service-account-token-expiration-seconds"
	AnnotationIDPTokenExpiration            = "workload-identity.stackit.cloud/idp-token-expiration-seconds"
	AnnotationIDPTokenEndpoint              = "workload-identity.stackit.cloud/idp-token-endpoint"
	AnnotationFederatedTokenFile            = "workload-identity.stackit.cloud/federated-token-file"

	// Pod annotations
	AnnotationSkipWebhook = "workload-identity.stackit.cloud/skip-pod-identity-webhook"

	// Defaults
	DefaultAudience                      = "sts.accounts.stackit.cloud"
	DefaultServiceAccountTokenExpiration = "600"
	DefaultVolumeName                    = "stackit-workload-identity"
	DefaultMountPath                     = "/var/run/secrets/stackit.cloud/serviceaccount"
	DefaultTokenPath                     = "token"

	// Environment variables
	EnvFederatedTokenFile        = "STACKIT_FEDERATED_TOKEN_FILE"
	EnvServiceAccountEmail       = "STACKIT_SERVICE_ACCOUNT_EMAIL"
	EnvIDPTokenEndpoint          = "STACKIT_IDP_TOKEN_ENDPOINT"
	EnvIDPTokenExpirationSeconds = "STACKIT_IDP_TOKEN_EXPIRATION_SECONDS"
)

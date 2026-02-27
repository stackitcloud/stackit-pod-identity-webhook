package webhook

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"path/filepath"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// This struct implements a mutating webhook for Kubernetes Pods.
// The webhook configuration is not automatically generated, but is manually defined
// in the Helm chart template at:
// charts/stackit-pod-identity-webhook/templates/mutatingwebhookconfiguration.yaml
//
// The key parameters for this webhook are:
// - Path: /mutate-v1-pod
// - Resources: pods
// - Operations: CREATE

// PodMutator mutates Pods
type PodMutator struct {
	Client client.Reader
}

// SetupWithManager sets up the webhook with the Manager.
func (m *PodMutator) SetupWithManager(mgr ctrl.Manager) error {
	if m.Client == nil {
		m.Client = mgr.GetClient()
	}

	return ctrl.NewWebhookManagedBy(mgr, &corev1.Pod{}).
		WithDefaulter(m).
		Complete()
}

var _ admission.Defaulter[*corev1.Pod] = &PodMutator{}

// ValidatedAnnotations holds the parsed and validated values from the ServiceAccount annotations.
type ValidatedAnnotations struct {
	ServiceAccountEmail           string
	Audience                      string
	ServiceAccountTokenExpiration int64
	IDPTokenExpiration            string
	IDPTokenEndpoint              string
	FederatedTokenFile            string
}

// Default implements admission.Defaulter so a webhook will be registered for the type
func (m *PodMutator) Default(ctx context.Context, pod *corev1.Pod) error {
	logger := log.FromContext(ctx).WithValues("pod", types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace})

	if (pod.Annotations != nil && pod.Annotations[AnnotationSkipWebhook] == "true") ||
		(pod.Labels != nil && pod.Labels[AnnotationSkipWebhook] == "true") {
		logger.Info("skipping webhook due to pod annotation or label")
		return nil
	}

	saName := pod.Spec.ServiceAccountName
	if saName == "" {
		saName = "default"
	}

	sa := &corev1.ServiceAccount{}
	err := m.Client.Get(ctx, types.NamespacedName{Name: saName, Namespace: pod.Namespace}, sa)
	if err != nil {
		logger.Error(err, "failed to fetch ServiceAccount", "sa", saName)
		return fmt.Errorf("failed to fetch ServiceAccount %s/%s: %w", pod.Namespace, saName, err)
	}

	if sa.Annotations != nil && sa.Annotations[AnnotationSkipWebhook] == "true" {
		logger.Info("skipping webhook due to service account annotation", "sa", saName)
		return nil
	}

	saEmail := sa.Annotations[AnnotationServiceAccountEmail]
	if saEmail == "" {
		logger.Info("workload identity not enabled for ServiceAccount", "sa", saName)
		return nil
	}

	logger.Info("mutating pod", "sa", saName, "saEmail", saEmail)

	// Mutate the pod
	err = m.HandleMutatePod(pod, sa, saEmail)
	if err != nil {
		logger.Error(err, "failed to mutate pod")
		return err
	}

	return nil
}

// validateAndParseAnnotations checks if all relevant annotations on the ServiceAccount are valid and returns their parsed values.
func validateAndParseAnnotations(annotations map[string]string) (ValidatedAnnotations, error) {
	var errs error
	parsedConfig := ValidatedAnnotations{
		Audience:           annotations[AnnotationAudience],
		IDPTokenExpiration: annotations[AnnotationIDPTokenExpiration],
		IDPTokenEndpoint:   annotations[AnnotationIDPTokenEndpoint],
		FederatedTokenFile: annotations[AnnotationFederatedTokenFile],
	}

	if parsedConfig.Audience == "" {
		parsedConfig.Audience = DefaultAudience
	}

	if email, ok := annotations[AnnotationServiceAccountEmail]; ok && email != "" {
		if _, err := mail.ParseAddress(email); err != nil {
			errs = errors.Join(errs, fmt.Errorf("invalid %q: %w", AnnotationServiceAccountEmail, err))
		}
		parsedConfig.ServiceAccountEmail = email
	}

	expirationSecondsStr := annotations[AnnotationServiceAccountTokenExpiration]
	if expirationSecondsStr == "" {
		expirationSecondsStr = DefaultServiceAccountTokenExpiration
	}
	if expiration, err := strconv.ParseInt(expirationSecondsStr, 10, 64); err != nil {
		errs = errors.Join(errs, fmt.Errorf("invalid %q: %w", AnnotationServiceAccountTokenExpiration, err))
	} else {
		parsedConfig.ServiceAccountTokenExpiration = expiration
	}

	if parsedConfig.IDPTokenExpiration != "" {
		if _, err := strconv.ParseInt(parsedConfig.IDPTokenExpiration, 10, 64); err != nil {
			errs = errors.Join(errs, fmt.Errorf("invalid %q: %w", AnnotationIDPTokenExpiration, err))
		}
	}

	if parsedConfig.IDPTokenEndpoint != "" {
		if _, err := url.ParseRequestURI(parsedConfig.IDPTokenEndpoint); err != nil {
			errs = errors.Join(errs, fmt.Errorf("invalid %q: %w", AnnotationIDPTokenEndpoint, err))
		}
	}

	return parsedConfig, errs
}

// HandleMutatePod applies the mutation logic to the pod spec
func (m *PodMutator) HandleMutatePod(pod *corev1.Pod, sa *corev1.ServiceAccount, saEmail string) error {
	parsedConfig, err := validateAndParseAnnotations(sa.Annotations)
	if err != nil {
		return err
	}

	volume, mountPath := createProjectedVolume(parsedConfig)

	// Check if volume in **podspec** already exists
	for _, v := range pod.Spec.Volumes {
		if v.Name == DefaultVolumeName {
			return fmt.Errorf("volume with name %q already exists in pod spec; please remove it to allow the webhook to inject the projected service account token volume", DefaultVolumeName)
		}
	}

	pod.Spec.Volumes = append(pod.Spec.Volumes, volume)

	envVars := prepareEnvVars(parsedConfig)

	for i := range pod.Spec.Containers {
		err := mutateContainer(&pod.Spec.Containers[i], envVars, mountPath)
		if err != nil {
			return fmt.Errorf("failed to mutate container %q: %w", pod.Spec.Containers[i].Name, err)
		}
	}
	for i := range pod.Spec.InitContainers {
		err := mutateContainer(&pod.Spec.InitContainers[i], envVars, mountPath)
		if err != nil {
			return fmt.Errorf("failed to mutate init container %q: %w", pod.Spec.InitContainers[i].Name, err)
		}
	}
	return nil
}

// createProjectedVolume builds the projected service account token volume based on validated annotations.
func createProjectedVolume(validated ValidatedAnnotations) (corev1.Volume, string) {
	mountPath := DefaultMountPath
	tokenPath := DefaultTokenPath
	if validated.FederatedTokenFile != "" {
		mountPath = filepath.Dir(validated.FederatedTokenFile)
		tokenPath = filepath.Base(validated.FederatedTokenFile)
	}

	volume := corev1.Volume{
		Name: DefaultVolumeName,
		VolumeSource: corev1.VolumeSource{
			Projected: &corev1.ProjectedVolumeSource{
				Sources: []corev1.VolumeProjection{
					{
						ServiceAccountToken: &corev1.ServiceAccountTokenProjection{
							Audience:          validated.Audience,
							ExpirationSeconds: &validated.ServiceAccountTokenExpiration,
							Path:              tokenPath,
						},
					},
				},
			},
		},
	}
	return volume, mountPath
}

// prepareEnvVars builds the slice of environment variables to be injected into the containers.
func prepareEnvVars(validated ValidatedAnnotations) []corev1.EnvVar {
	envVars := []corev1.EnvVar{
		{
			Name:  EnvServiceAccountEmail,
			Value: validated.ServiceAccountEmail,
		},
	}

	if validated.FederatedTokenFile != "" {
		envVars = append(envVars, corev1.EnvVar{
			Name:  EnvFederatedTokenFile,
			Value: validated.FederatedTokenFile,
		})
	}

	if validated.IDPTokenEndpoint != "" {
		envVars = append(envVars, corev1.EnvVar{
			Name:  EnvIDPTokenEndpoint,
			Value: validated.IDPTokenEndpoint,
		})
	}

	if validated.IDPTokenExpiration != "" {
		envVars = append(envVars, corev1.EnvVar{
			Name:  EnvIDPTokenExpirationSeconds,
			Value: validated.IDPTokenExpiration,
		})
	}
	return envVars
}

// mutateContainer injects the required volume mount and environment variables into a container.
func mutateContainer(c *corev1.Container, envVars []corev1.EnvVar, mountPath string) error {
	var err error
	c.VolumeMounts, err = addVolumeMount(c.VolumeMounts, mountPath)
	if err != nil {
		return err
	}
	c.Env, err = addEnvVars(c.Env, envVars)
	return err
}

// addVolumeMount adds a volume mount to the given slice if it doesn't already exist.
func addVolumeMount(volumeMounts []corev1.VolumeMount, mountPath string) ([]corev1.VolumeMount, error) {
	for _, vm := range volumeMounts {
		if vm.Name == DefaultVolumeName {
			return nil, fmt.Errorf("volumeMount with name %q already exists", DefaultVolumeName)
		}
	}
	return append(volumeMounts, corev1.VolumeMount{
		Name:      DefaultVolumeName,
		MountPath: mountPath,
		ReadOnly:  true,
	}), nil
}

// addEnvVars adds environment variables to the given slice if they don't already exist.
func addEnvVars(envs []corev1.EnvVar, newEnvs []corev1.EnvVar) ([]corev1.EnvVar, error) {
	for _, ev := range newEnvs {
		for _, existingEv := range envs {
			if existingEv.Name == ev.Name {
				return nil, fmt.Errorf("env var with name %q already exists", ev.Name)
			}
		}
		envs = append(envs, ev)
	}
	return envs, nil
}

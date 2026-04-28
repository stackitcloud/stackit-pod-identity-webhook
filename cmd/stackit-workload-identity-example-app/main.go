// Package main provides a simple example application that demonstrates the use of STACKIT Workload Identity.
// It uses the STACKIT Go SDK to interact with the SKE API, relying on the identity injected
// by the stackit-pod-identity-webhook for authentication.
// Getting the provider options does not require any permissions to be assigned to the ServiceAccount.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/stackitcloud/stackit-sdk-go/core/config"
	ske "github.com/stackitcloud/stackit-sdk-go/services/ske/v2api"
)

const defaultRegion = "eu01"

func main() {
	if err := run(); err != nil {
		slog.Error("Application failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	region := os.Getenv("STACKIT_REGION")
	if region == "" {
		region = defaultRegion
	}

	var opts []config.ConfigurationOption
	if endpoint := os.Getenv("STACKIT_SKE_ENDPOINT"); endpoint != "" {
		slog.Info("Using custom SKE endpoint", "endpoint", endpoint)
		opts = append(opts, config.WithEndpoint(endpoint))
	}

	// Create a new API client that uses default authentication and configuration
	skeClient, err := ske.NewAPIClient(opts...)
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	slog.Info("Fetching SKE options", "region", region)
	getOptionsResp, err := skeClient.DefaultAPI.ListProviderOptions(context.Background(), region).Execute()
	if err != nil {
		return fmt.Errorf("calling ListProviderOptions: %w", err)
	}

	slog.Info("Authentication successful, API call succeeded")

	availableVersions := getOptionsResp.KubernetesVersions
	if len(availableVersions) == 0 {
		slog.Warn("No Kubernetes versions found", "region", region)
	}

	return nil
}

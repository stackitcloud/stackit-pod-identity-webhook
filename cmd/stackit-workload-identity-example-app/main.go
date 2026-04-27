// Package main provides a simple example application that demonstrates the use of STACKIT Workload Identity.
// It uses the STACKIT Go SDK to interact with the SKE API, relying on the identity injected
// by the stackit-pod-identity-webhook for authentication.
// Getting the provider options does not require any permissions to be assigned to the ServiceAccount.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/stackitcloud/stackit-sdk-go/core/config"
	ske "github.com/stackitcloud/stackit-sdk-go/services/ske/v2api"
)

const defaultRegion = "eu01"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
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
		fmt.Printf("Using custom SKE endpoint: %s\n", endpoint)
		opts = append(opts, config.WithEndpoint(endpoint))
	}

	// Create a new API client, that uses default authentication and configuration
	skeClient, err := ske.NewAPIClient(opts...)
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	fmt.Printf("Fetching SKE options for region %q...\n", region)
	getOptionsResp, err := skeClient.DefaultAPI.ListProviderOptions(context.Background(), region).Execute()
	if err != nil {
		return fmt.Errorf("calling ListProviderOptions: %w", err)
	}

	fmt.Println("Authentication successful, API call succeeded")

	availableVersions := getOptionsResp.KubernetesVersions
	if len(availableVersions) == 0 {
		fmt.Printf("WARNING: No Kubernetes versions found for region %q\n", region)
	}

	return nil
}

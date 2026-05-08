// Package main provides a simple example application that demonstrates the use of STACKIT Workload Identity.
// It uses the STACKIT Go SDK to interact with the SKE API, relying on the identity injected
// by the stackit-pod-identity-webhook for authentication.
// Getting the public IP ranges does not require any permissions to be assigned to the ServiceAccount.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/stackitcloud/stackit-sdk-go/core/config"
	iaas "github.com/stackitcloud/stackit-sdk-go/services/iaas/v2api"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		slog.Error("Application failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	var opts []config.ConfigurationOption
	if endpoint := os.Getenv("STACKIT_IAAS_API_ENDPOINT"); endpoint != "" {
		slog.Info("Using custom IaaS API endpoint", "endpoint", endpoint)
		opts = append(opts, config.WithEndpoint(endpoint))
	}

	// Create a new API client that uses default authentication and configuration
	iaasClient, err := iaas.NewAPIClient(opts...)
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	slog.Info("Fetching public IP ranges")

	publicIpRangesResponse, err := iaasClient.DefaultAPI.ListPublicIPRanges(ctx).Execute()

	if err != nil {
		return fmt.Errorf("calling ListPublicIPRanges: %w", err)
	}

	slog.Info("Authentication successful, API call succeeded")

	publicIpRanges := publicIpRangesResponse.Items

	if len(publicIpRanges) == 0 {
		slog.Warn("No public IP ranges found. There might be a problem with the autentication.")
	}

	return nil
}

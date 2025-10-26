// A Dagger module for K3s cluster management and ArgoCD deployment
//
// This module uses the k3s daggerverse module to:
// - Create a K3s Kubernetes cluster
// - Deploy ArgoCD to the cluster
// - Verify the deployment
//
// All operations are containerized via Dagger.

package main

import (
	"context"
	"fmt"
	"time"
)

type Daggerk3 struct{}

const (
	argoCDVersion = "v2.13.2"
	argoCDNS      = "argocd"
)

// IntegrationTest runs a complete integration test:
// - Creates a K3s cluster using the k3s daggerverse module
// - Deploys ArgoCD to the cluster
// - Waits for ArgoCD to be ready
// - Returns the deployment status
func (m *Daggerk3) IntegrationTest(
	ctx context.Context,
	// Name of the test cluster
	// +optional
	// +default="argocd-test"
	clusterName string,
) (string, error) {
	if clusterName == "" {
		clusterName = "argocd-test"
	}

	// Create K3s cluster using the daggerverse module
	k3s := dag.K3S(clusterName)

	// Start the server explicitly (like in the example)
	server := k3s.Server()
	_, err := server.Start(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to start k3s server: %w", err)
	}

	// Now use kubectl with the running server
	// Create argocd namespace
	createCmd := fmt.Sprintf("create namespace %s", argoCDNS)
	_, err = k3s.Kubectl(createCmd).Sync(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create namespace: %w", err)
	}

	// Install ArgoCD
	manifestURL := fmt.Sprintf("https://raw.githubusercontent.com/argoproj/argo-cd/%s/manifests/install.yaml", argoCDVersion)
	applyCmd := fmt.Sprintf("apply -n %s -f %s", argoCDNS, manifestURL)
	deployOutput, err := k3s.Kubectl(applyCmd).Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to install ArgoCD: %w", err)
	}

	// Wait for ArgoCD deployments to be ready
	waitCmd := fmt.Sprintf("wait --for=condition=available --timeout=300s --all deployments -n %s", argoCDNS)
	waitOutput, err := k3s.Kubectl(waitCmd).Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("ArgoCD failed to become ready: %w", err)
	}

	// Get final status
	statusCmd := fmt.Sprintf("get pods -n %s", argoCDNS)
	status, err := k3s.Kubectl(statusCmd).Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get ArgoCD status: %w", err)
	}

	result := fmt.Sprintf(`Integration Test Successful!

Cluster: %s
Timestamp: %s

=== Deployment Output ===
%s

=== Wait Output ===
%s

=== ArgoCD Pods ===
%s`,
		clusterName, time.Now().Format(time.RFC3339), deployOutput, waitOutput, status)

	return result, nil
}


package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/tiny-systems/module/pkg/resource"
	"github.com/tiny-systems/module/pkg/utils"
	"k8s.io/client-go/rest"
)

// PortForwardClient implements utils.ClientInterface using kubectl port-forward.
// This is used by the desktop client to access services inside the Kubernetes cluster.
type PortForwardClient struct {
	config    *rest.Config
	namespace string

	mu sync.Mutex
	pf *resource.PortForwarder
}

// NewPortForwardClient creates a new PortForwardClient
func NewPortForwardClient(config *rest.Config, namespace string) *PortForwardClient {
	return &PortForwardClient{
		config:    config,
		namespace: namespace,
	}
}

// GetForwardedAddress implements utils.ClientInterface.
// It creates a port-forward to the specified service and returns localhost:port.
func (c *PortForwardClient) GetForwardedAddress(ctx context.Context, req utils.PortForwardRequest, alias string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Create port forwarder if needed
	if c.pf == nil {
		namespace := req.Namespace
		if namespace == "" {
			namespace = c.namespace
		}

		pf, err := resource.CreatePortForwarderFromConfig(c.config, namespace)
		if err != nil {
			return "", fmt.Errorf("failed to create port forwarder: %w", err)
		}
		c.pf = pf
	}

	// Forward the service port
	addr, err := c.pf.ForwardService(ctx, req.ServiceName, req.Port)
	if err != nil {
		return "", fmt.Errorf("failed to forward service %s:%d: %w", req.ServiceName, req.Port, err)
	}

	return addr, nil
}

// Close closes the port forwarder and releases resources
func (c *PortForwardClient) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.pf != nil {
		c.pf.StopAll()
		c.pf = nil
	}
}

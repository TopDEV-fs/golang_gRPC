// Package productcatalog exposes a thin public facade over the internal service
// container. External consumers (e.g. integration tests, other services in the
// same binary) should wire through this package rather than reaching into
// internal/services directly. This preserves the Go visibility boundary while
// providing a stable composition point.
package productcatalog

import (
	"cloud.google.com/go/spanner"
	"github.com/example/product-catalog-service/internal/services"
)

// Service is the public facade that wraps the application Container.
type Service struct {
	container *services.Container
}

// New creates a Service, wires all dependencies, and returns it ready for use.
// It is the recommended entry point for embedding this service in a larger binary.
func New(client *spanner.Client) *Service {
	return &Service{container: services.BuildContainer(client)}
}

// Container returns the underlying application Container, giving callers access
// to individual command and query handlers.
func (s *Service) Container() *services.Container {
	return s.container
}

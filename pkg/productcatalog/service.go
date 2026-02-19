package productcatalog

import (
	"cloud.google.com/go/spanner"
	"github.com/example/product-catalog-service/internal/services"
)

// Service is a small public facade exposing command/query composition.
type Service struct {
	container *services.Container
}

func New(client *spanner.Client) *Service {
	return &Service{container: services.BuildContainer(client)}
}

func (s *Service) Container() *services.Container {
	return s.container
}

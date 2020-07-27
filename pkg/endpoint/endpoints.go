package endpoint

import (
	"context"

	"github.com/PedPet/api-gateway/config"
	userSvcService "github.com/PedPet/user/pkg/service"
	"google.golang.org/grpc"
)

// Endpoints is a collection of all other
type Endpoints struct {
	User User
}

// grpc service connection function type for return type in connection functions
type grpcSvcConnFunc func() (userSvcService.User, *grpc.ClientConn, error)

// MakeEndpoints creates all endpoints
func MakeEndpoints(ctx context.Context, s *config.Settings) Endpoints {
	return Endpoints{
		User: makeUserEndpoints(ctx, s.User),
	}
}

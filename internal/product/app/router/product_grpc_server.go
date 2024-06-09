package router

import (
	"github.com/google/wire"
	"github.com/namdam97/go-coffee/internal/product/usecases/products"
	"github.com/namdam97/go-coffee/proto/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var _ gen.ProductServiceServer = (*productGRPCServer)(nil)

var ProductGRPCServerSet = wire.NewSet(NewProductGRPCServer)

type productGRPCServer struct {
	gen.UnimplementedProductServiceServer
	uc products.UseCase
}

func NewProductGRPCServer(
	grpcServer *grpc.Server,
	uc products.UseCase,
) gen.ProductServiceServer {
	svc := productGRPCServer{
		uc: uc,
	}

	gen.RegisterProductServiceServer(grpcServer, &svc)

	reflection.Register(grpcServer)

	return &svc
}

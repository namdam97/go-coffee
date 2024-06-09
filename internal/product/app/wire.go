//go:build wireinject
// +build wireinject

package app

import (
	"github.com/google/wire"
	"github.com/namdam97/go-coffee/cmd/product/config"
	"github.com/namdam97/go-coffee/internal/product/app/router"
	"github.com/namdam97/go-coffee/internal/product/infras/repo"
	productsUC "github.com/namdam97/go-coffee/internal/product/usecases/products"
	"google.golang.org/grpc"
)

func InitApp(
	cfg *config.Config,
	grpcServer *grpc.Server,
) (*App, error) {
	panic(wire.Build(
		New,
		router.ProductGRPCServerSet,
		repo.RepositorySet,
		productsUC.UseCaseSet,
	))
}

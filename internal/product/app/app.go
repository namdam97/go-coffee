package app

import (
	"github.com/namdam97/go-coffee/cmd/product/config"
	productUC "github.com/namdam97/go-coffee/internal/product/usecases/products"
	"github.com/namdam97/go-coffee/proto/gen"
)

type App struct {
	Cfg               *config.Config
	UC                productUC.UseCase
	ProductGRPCServer gen.ProductServiceServer
}

func New(
	cfg *config.Config,
	uc productUC.UseCase,
	productGRPCServer gen.ProductServiceServer,
) *App {
	return &App{
		Cfg:               cfg,
		UC:                uc,
		ProductGRPCServer: productGRPCServer,
	}
}

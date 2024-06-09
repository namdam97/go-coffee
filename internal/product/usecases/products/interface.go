package products

import (
	"context"

	"github.com/namdam97/go-coffee/internal/product/domain"
)

type UseCase interface {
	GetItemTypes(context.Context) ([]*domain.ItemTypeDto, error)
	GetItemsByType(context.Context, string) ([]*domain.ItemDto, error)
}

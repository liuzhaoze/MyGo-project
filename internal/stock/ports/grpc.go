package ports

import (
	"context"
	"github.com/liuzhaoze/MyGo-project/common/genproto/stockpb"
	"github.com/liuzhaoze/MyGo-project/common/tracing"
	"github.com/liuzhaoze/MyGo-project/stock/app"
	"github.com/liuzhaoze/MyGo-project/stock/app/query"
	"github.com/liuzhaoze/MyGo-project/stock/converter"
)

type GRPCServer struct {
	app app.Application
}

func NewGRPCServer(app app.Application) *GRPCServer {
	return &GRPCServer{app: app}
}

func (G GRPCServer) GetItems(ctx context.Context, request *stockpb.GetItemsRequest) (*stockpb.GetItemsResponse, error) {
	_, span := tracing.Start(ctx, "get_items")
	defer span.End()

	items, err := G.app.Queries.GetItems.Handle(ctx, query.GetItems{ItemIDs: request.ItemIDs})
	if err != nil {
		return nil, err
	}
	return &stockpb.GetItemsResponse{Items: items}, nil
}

func (G GRPCServer) CheckIfItemsInStock(ctx context.Context, request *stockpb.CheckIfItemsInStockRequest) (*stockpb.CheckIfItemsInStockResponse, error) {
	_, span := tracing.Start(ctx, "check_if_items_in_stock")
	defer span.End()

	items, err := G.app.Queries.CheckIfItemsInStock.Handle(ctx, query.CheckIfItemsInStock{
		Items: converter.NewItemWithQuantityConverter().ProtosToEntities(request.Items),
	})
	if err != nil {
		return nil, err
	}
	return &stockpb.CheckIfItemsInStockResponse{
		InStock: 1,
		Items:   converter.NewItemConverter().EntitiesToProtos(items),
	}, nil
}

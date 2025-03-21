package ports

import (
	"context"
	"github.com/liuzhaoze/MyGo-project/order/converter"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/liuzhaoze/MyGo-project/common/genproto/orderpb"
	"github.com/liuzhaoze/MyGo-project/order/app"
	"github.com/liuzhaoze/MyGo-project/order/app/command"
	"github.com/liuzhaoze/MyGo-project/order/app/query"
	domain "github.com/liuzhaoze/MyGo-project/order/domain/order"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GRPCServer struct {
	app app.Application
}

func NewGRPCServer(app app.Application) *GRPCServer {
	return &GRPCServer{app: app}
}

func (G GRPCServer) CreateOrder(ctx context.Context, request *orderpb.CreateOrderRequest) (*emptypb.Empty, error) {
	_, err := G.app.Commands.CreateOrder.Handle(ctx, command.CreateOrder{
		CustomerID: request.CustomerID,
		Items:      converter.NewItemWithQuantityConverter().ProtosToEntities(request.Items),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &empty.Empty{}, nil
}

func (G GRPCServer) GetOrder(ctx context.Context, request *orderpb.GetOrderRequest) (*orderpb.Order, error) {
	o, err := G.app.Queries.GetCustomerOrder.Handle(ctx, query.GetCustomerOrder{
		CustomerID: request.CustomerID,
		OrderID:    request.OrderID,
	})
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return converter.NewOrderConverter().EntityToProto(o), nil
}

func (G GRPCServer) UpdateOrder(ctx context.Context, request *orderpb.Order) (*emptypb.Empty, error) {
	o, err := domain.NewOrder(
		request.ID,
		request.CustomerID,
		request.Status,
		request.PaymentLink,
		converter.NewItemConverter().ProtosToEntities(request.Items),
	)
	if err != nil {
		err = status.Error(codes.Internal, err.Error())
		return nil, err
	}

	_, err = G.app.Commands.UpdateOrder.Handle(ctx, command.UpdateOrder{
		Order: o,
		UpdateFn: func(ctx context.Context, order *domain.Order) (*domain.Order, error) {
			if err := order.UpdateStatusTo(request.Status); err != nil {
				return nil, err
			}
			if err := order.UpdatePaymentLink(request.PaymentLink); err != nil {
				return nil, err
			}
			if err := order.UpdateItems(converter.NewItemConverter().ProtosToEntities(request.Items)); err != nil {
				return nil, err
			}
			return order, nil
		},
	})
	return nil, err
}

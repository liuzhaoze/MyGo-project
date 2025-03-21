package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/liuzhaoze/MyGo-project/common"
	client "github.com/liuzhaoze/MyGo-project/common/client/order"
	"github.com/liuzhaoze/MyGo-project/common/consts"
	"github.com/liuzhaoze/MyGo-project/common/handler/errors"
	"github.com/liuzhaoze/MyGo-project/order/app"
	"github.com/liuzhaoze/MyGo-project/order/app/command"
	"github.com/liuzhaoze/MyGo-project/order/app/dto"
	"github.com/liuzhaoze/MyGo-project/order/app/query"
	"github.com/liuzhaoze/MyGo-project/order/converter"
)

type HTTPServer struct {
	common.BaseResponse
	app app.Application
}

func (H HTTPServer) PostCustomerCustomerIdOrders(c *gin.Context, customerID string) {
	var (
		req  client.CreateOrderRequest
		err  error
		resp dto.CreateOrderResponse
	)
	defer func() {
		H.Response(c, err, &resp)
	}()

	if err = c.ShouldBindJSON(&req); err != nil {
		err = errors.NewWithError(consts.ErrnoBindRequestError, err)
		return
	}
	if err = H.validate(req); err != nil {
		err = errors.NewWithError(consts.ErrorRequestValidateError, err)
		return
	}

	r, err := H.app.Commands.CreateOrder.Handle(c.Request.Context(), command.CreateOrder{
		CustomerID: customerID,
		Items:      converter.NewItemWithQuantityConverter().ClientsToEntities(req.Items),
	})
	if err != nil {
		return
	}

	resp = dto.CreateOrderResponse{
		OrderID:     r.OrderID,
		CustomerID:  req.CustomerId,
		RedirectURL: fmt.Sprintf("http://localhost:8282/success?customerID=%s&orderID=%s", req.CustomerId, r.OrderID),
	}
}

func (H HTTPServer) GetCustomerCustomerIdOrdersOrderId(c *gin.Context, customerID string, orderID string) {
	var (
		err  error
		resp interface{}
	)
	defer func() {
		H.Response(c, err, &resp)
	}()
	o, err := H.app.Queries.GetCustomerOrder.Handle(c.Request.Context(), query.GetCustomerOrder{
		CustomerID: customerID,
		OrderID:    orderID,
	})

	if err != nil {
		return
	}

	resp = converter.NewOrderConverter().EntityToClient(o)
}

func (H HTTPServer) validate(req client.CreateOrderRequest) error {
	for _, v := range req.Items {
		if v.Quantity <= 0 {
			return fmt.Errorf("quantity must be positive, got %d from %s", v.Quantity, v.Id)
		}
	}
	return nil
}

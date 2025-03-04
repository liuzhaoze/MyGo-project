package main

import (
	"github.com/gin-gonic/gin"
	"github.com/liuzhaoze/MyGo-project/common/genproto/orderpb"
	"github.com/liuzhaoze/MyGo-project/order/app"
	"github.com/liuzhaoze/MyGo-project/order/app/command"
	"github.com/liuzhaoze/MyGo-project/order/app/query"
	"net/http"
)

type HTTPServer struct {
	app app.Application
}

func (H HTTPServer) PostCustomerCustomerIDOrders(c *gin.Context, customerID string) {
	var req orderpb.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	r, err := H.app.Commands.CreateOrder.Handle(c, command.CreateOrder{
		CustomerID: customerID,
		Items:      req.Items,
	})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":     "success",
		"customer_id": req.CustomerID,
		"order_id":    r.OrderID,
	})
}

func (H HTTPServer) GetCustomerCustomerIDOrdersOrderID(c *gin.Context, customerID string, orderID string) {
	o, err := H.app.Queries.GetCustomerOrder.Handle(c, query.GetCustomerOrder{
		CustomerID: customerID,
		OrderID:    orderID,
	})

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success", "data": o})

}

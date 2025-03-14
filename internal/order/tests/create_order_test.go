package tests

import (
	"context"
	"fmt"
	serverWrapper "github.com/liuzhaoze/MyGo-project/common/client/order"
	_ "github.com/liuzhaoze/MyGo-project/common/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"testing"
)

var (
	ctx    = context.Background()
	server = fmt.Sprintf("http://%s/api", viper.GetString("order.http-addr"))
)

func TestMain(m *testing.M) {
	before()
	m.Run()
}

func before() {
	log.Printf("server=%s", server)
}

func TestCreateOrder_success(t *testing.T) {
	customerID := "123"
	resp := getResponse(t, customerID, serverWrapper.PostCustomerCustomerIdOrdersJSONRequestBody{
		CustomerId: customerID,
		Items: []serverWrapper.ItemWithQuantity{
			{
				Id:       "test-item-1",
				Quantity: int32(1),
			},
			{
				Id:       "test-item-2",
				Quantity: int32(2),
			},
		},
	})

	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, 0, resp.JSON200.ErrorCode)
}

func TestCreateOrder_empty_items(t *testing.T) {
	customerID := "123"
	resp := getResponse(t, customerID, serverWrapper.PostCustomerCustomerIdOrdersJSONRequestBody{
		CustomerId: customerID,
		Items:      nil,
	})

	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, 2, resp.JSON200.ErrorCode)
}

func getResponse(t *testing.T, customerID string, body serverWrapper.PostCustomerCustomerIdOrdersJSONRequestBody) *serverWrapper.PostCustomerCustomerIdOrdersResponse {
	t.Helper()
	client, err := serverWrapper.NewClientWithResponses(server)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.PostCustomerCustomerIdOrdersWithResponse(ctx, customerID, body)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

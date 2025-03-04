package app

import "github.com/liuzhaoze/MyGo-project/order/app/query"

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
}

type Queries struct {
	GetCustomerOrder query.GetCustomerOrderHandler
}

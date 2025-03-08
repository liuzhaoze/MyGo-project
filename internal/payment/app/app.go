package app

import "github.com/liuzhaoze/MyGo-project/payment/app/command"

type Application struct {
	Commands Commands
}

type Commands struct {
	CreatePayment command.CreatePaymentHandler
}

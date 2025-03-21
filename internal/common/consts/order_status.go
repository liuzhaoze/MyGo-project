package consts

// OrderStatus 不用是因为替换需要改动的地方太多了
type OrderStatus string

const (
	OrderStatusPending           = "pending"
	OrderStatusWaitingForPayment = "waiting_for_payment"
	OrderStatusPaid              = "paid"
	OrderStatusReady             = "ready"
)

package order

import (
	"errors"
	"fmt"
	"github.com/liuzhaoze/MyGo-project/common/consts"
	"github.com/liuzhaoze/MyGo-project/order/entity"
	"slices"
)

// Order
// Aggregate
type Order struct {
	ID          string
	CustomerID  string
	Status      string
	PaymentLink string
	Items       []*entity.Item
}

func NewOrder(id, customerID, status, paymentLink string, items []*entity.Item) (*Order, error) {
	if id == "" {
		return nil, errors.New("empty id")
	}

	if customerID == "" {
		return nil, errors.New("empty customerID")
	}

	if status == "" {
		return nil, errors.New("empty status")
	}

	if items == nil {
		return nil, errors.New("empty items")
	}

	return &Order{
		ID:          id,
		CustomerID:  customerID,
		Status:      consts.OrderStatusPending,
		PaymentLink: paymentLink,
		Items:       items,
	}, nil
}

func NewPendingOrder(customerID string, items []*entity.Item) (*Order, error) {
	if customerID == "" {
		return nil, errors.New("empty customerID")
	}

	if items == nil {
		return nil, errors.New("empty items")
	}

	return &Order{
		CustomerID: customerID,
		Status:     "pending",
		Items:      items,
	}, nil
}

func (o *Order) UpdateStatusTo(targetStatus string) error {
	if !o.isValidStatusTransition(targetStatus) {
		return fmt.Errorf("cannot update status from %s to %s", o.Status, targetStatus)
	}
	o.Status = targetStatus
	return nil
}

func (o *Order) isValidStatusTransition(status string) bool {
	switch o.Status {
	default:
		return false
	case consts.OrderStatusPending:
		return slices.Contains([]string{consts.OrderStatusWaitingForPayment}, status)
	case consts.OrderStatusWaitingForPayment:
		return slices.Contains([]string{consts.OrderStatusPaid}, status)
	case consts.OrderStatusPaid:
		return slices.Contains([]string{consts.OrderStatusReady}, status)
	}
}

func (o *Order) UpdatePaymentLink(link string) error {
	o.PaymentLink = link
	return nil
}

func (o *Order) UpdateItems(items []*entity.Item) error {
	o.Items = items
	return nil
}

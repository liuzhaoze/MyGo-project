package order

import (
	"errors"
	"fmt"
	"github.com/liuzhaoze/MyGo-project/order/entity"

	"github.com/stripe/stripe-go/v81"
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
		Status:      status,
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

func (o *Order) IsPaid() error {
	if o.Status == string(stripe.CheckoutSessionPaymentStatusPaid) {
		return nil
	}
	return fmt.Errorf("order is not paid, order id=%s, status=%s", o.ID, o.Status)
}

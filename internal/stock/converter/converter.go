package converter

import (
	"github.com/liuzhaoze/MyGo-project/common/genproto/orderpb"
	"github.com/liuzhaoze/MyGo-project/stock/entity"
)

type OrderConverter struct {
}

type ItemConverter struct {
}

type ItemWithQuantityConverter struct {
}

func (c *ItemWithQuantityConverter) EntitiesToProtos(items []*entity.ItemWithQuantity) (res []*orderpb.ItemWithQuantity) {
	for _, i := range items {
		res = append(res, c.EntityToProto(i))
	}
	return
}

func (c *ItemWithQuantityConverter) EntityToProto(i *entity.ItemWithQuantity) *orderpb.ItemWithQuantity {
	return &orderpb.ItemWithQuantity{
		ID:       i.ID,
		Quantity: i.Quantity,
	}
}

func (c *ItemWithQuantityConverter) ProtosToEntities(items []*orderpb.ItemWithQuantity) (res []*entity.ItemWithQuantity) {
	for _, i := range items {
		res = append(res, c.ProtoToEntity(i))
	}
	return
}

func (c *ItemWithQuantityConverter) ProtoToEntity(i *orderpb.ItemWithQuantity) *entity.ItemWithQuantity {
	return &entity.ItemWithQuantity{
		ID:       i.ID,
		Quantity: i.Quantity,
	}
}

func (oc *OrderConverter) EntityToProto(o *entity.Order) *orderpb.Order {
	oc.check(o)
	return &orderpb.Order{
		ID:          o.ID,
		CustomerID:  o.CustomerID,
		Status:      o.Status,
		Items:       NewItemConverter().EntitiesToProtos(o.Items),
		PaymentLink: o.PaymentLink,
	}
}

func (oc *OrderConverter) ProtoToEntity(o *orderpb.Order) *entity.Order {
	oc.check(o)
	return &entity.Order{
		ID:          o.ID,
		CustomerID:  o.CustomerID,
		Status:      o.Status,
		PaymentLink: o.PaymentLink,
		Items:       NewItemConverter().ProtosToEntities(o.Items),
	}
}

func (oc *OrderConverter) check(o interface{}) {
	if o == nil {
		panic("cannot convert nil order")
	}
}

func (ic *ItemConverter) EntitiesToProtos(items []*entity.Item) (res []*orderpb.Item) {
	for _, i := range items {
		res = append(res, ic.EntityToProto(i))
	}
	return
}

func (ic *ItemConverter) ProtosToEntities(items []*orderpb.Item) (res []*entity.Item) {
	for _, i := range items {
		res = append(res, ic.ProtoToEntity(i))
	}
	return
}

func (ic *ItemConverter) EntityToProto(i *entity.Item) *orderpb.Item {
	return &orderpb.Item{
		ID:       i.ID,
		Name:     i.Name,
		Quantity: i.Quantity,
		PriceID:  i.PriceID,
	}
}

func (ic *ItemConverter) ProtoToEntity(i *orderpb.Item) *entity.Item {
	return &entity.Item{
		ID:       i.ID,
		Name:     i.Name,
		Quantity: i.Quantity,
		PriceID:  i.PriceID,
	}
}

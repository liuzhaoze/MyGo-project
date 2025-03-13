package adapters

import (
	"context"
	_ "github.com/liuzhaoze/MyGo-project/common/config"
	domain "github.com/liuzhaoze/MyGo-project/order/domain/order"
	"github.com/liuzhaoze/MyGo-project/order/entity"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

var (
	dbName   = viper.GetString("mongo.db-name")
	collName = viper.GetString("mongo.coll-name")
)

type OrderRepositoryMongo struct {
	db *mongo.Client
}

func NewOrderRepositoryMongo(db *mongo.Client) *OrderRepositoryMongo {
	return &OrderRepositoryMongo{db: db}
}

func (r *OrderRepositoryMongo) collection() *mongo.Collection {
	return r.db.Database(dbName).Collection(collName)
}

type orderModel struct {
	MongoID     primitive.ObjectID `bson:"_id"`
	ID          string             `bson:"id"`
	CustomerID  string             `bson:"customer_id"`
	Status      string             `bson:"status"`
	PaymentLink string             `bson:"payment_link"`
	Items       []*entity.Item     `bson:"items"`
}

func (r *OrderRepositoryMongo) Create(ctx context.Context, order *domain.Order) (created *domain.Order, err error) {
	defer func() {
		r.logWithTag("Create", err, created)
	}()
	writeModel := r.marshalToModel(order)
	res, err := r.collection().InsertOne(ctx, writeModel)
	if err != nil {
		return nil, err
	}
	created = order
	created.ID = res.InsertedID.(primitive.ObjectID).Hex()
	return
}

func (r *OrderRepositoryMongo) logWithTag(tag string, err error, result interface{}) {
	l := logrus.WithFields(logrus.Fields{
		"tag":          "OrderRepositoryMongo",
		"operate_time": time.Now().Unix(),
		"err":          err,
		"info":         result,
	})
	if err != nil {
		l.Infof("%s fail", tag)
	} else {
		l.Infof("%s success", tag)
	}
}

func (r *OrderRepositoryMongo) Get(ctx context.Context, id, customerID string) (got *domain.Order, err error) {
	defer func() {
		r.logWithTag("Get", err, got)
	}()
	readModel := &orderModel{}
	mongoID, _ := primitive.ObjectIDFromHex(id)
	condition := bson.M{"_id": mongoID}
	if err = r.collection().FindOne(ctx, condition).Decode(readModel); err != nil {
		return
	}

	if readModel == nil {
		return nil, &domain.NotFoundError{OrderID: id}
	}
	return r.unmarshal(readModel), nil
}

// Update 先查找对应的 order ，然后 apply updateFn ，再写入回去
func (r *OrderRepositoryMongo) Update(
	ctx context.Context,
	order *domain.Order,
	updateFn func(context.Context, *domain.Order) (*domain.Order, error),
) (err error) {
	defer func() {
		r.logWithTag("Update", err, nil)
	}()
	if order == nil {
		panic("nil order")
	}

	// 启动 MongoDB 事务
	session, err := r.db.StartSession()
	if err != nil {
		return
	}
	defer session.EndSession(ctx)
	if err = session.StartTransaction(); err != nil {
		return err
	}
	defer func() {
		if err == nil {
			_ = session.CommitTransaction(ctx)
		} else {
			_ = session.AbortTransaction(ctx)
		}
	}()

	// inside transaction
	oldOrder, err := r.Get(ctx, order.ID, order.CustomerID)
	if err != nil {
		return
	}
	updatedOrder, err := updateFn(ctx, order)
	if err != nil {
		return
	}

	mongoID, _ := primitive.ObjectIDFromHex(oldOrder.ID)
	res, err := r.collection().UpdateOne(
		ctx,
		bson.M{"_id": mongoID, "customer_id": oldOrder.CustomerID},
		bson.M{"$set": bson.M{"status": updatedOrder.Status, "payment_link": updatedOrder.PaymentLink}},
	)
	if err != nil {
		return
	}

	r.logWithTag("Update", err, res)

	return
}

func (r *OrderRepositoryMongo) marshalToModel(order *domain.Order) *orderModel {
	return &orderModel{
		MongoID:     primitive.NewObjectID(),
		ID:          order.ID,
		CustomerID:  order.CustomerID,
		Status:      order.Status,
		PaymentLink: order.PaymentLink,
		Items:       order.Items,
	}
}

func (r *OrderRepositoryMongo) unmarshal(m *orderModel) *domain.Order {
	return &domain.Order{
		ID:          m.MongoID.Hex(),
		CustomerID:  m.CustomerID,
		Status:      m.Status,
		PaymentLink: m.PaymentLink,
		Items:       m.Items,
	}
}

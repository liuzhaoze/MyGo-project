package adapters

import (
	"context"
	"github.com/liuzhaoze/MyGo-project/stock/entity"
	"github.com/liuzhaoze/MyGo-project/stock/infrastructure/persistent"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MySQLStockRepository struct {
	db *persistent.MySQL
}

func NewMySQLStockRepository(db *persistent.MySQL) *MySQLStockRepository {
	return &MySQLStockRepository{db: db}
}

func (m MySQLStockRepository) GetItems(ctx context.Context, ids []string) ([]*entity.Item, error) {
	//TODO implement me
	panic("implement me")
}

func (m MySQLStockRepository) GetStock(ctx context.Context, ids []string) ([]*entity.ItemWithQuantity, error) {
	data, err := m.db.BatchGetStockByID(ctx, ids)
	if err != nil {
		return nil, errors.Wrap(err, "BatchGetStockByID error")
	}
	var result []*entity.ItemWithQuantity
	for _, d := range data {
		result = append(result, &entity.ItemWithQuantity{
			ID:       d.ProductID,
			Quantity: d.Quantity,
		})
	}
	return result, nil
}

func (m MySQLStockRepository) UpdateStock(
	ctx context.Context,
	data []*entity.ItemWithQuantity,
	updateFn func(
	ctx context.Context,
	existing []*entity.ItemWithQuantity,
	query []*entity.ItemWithQuantity,
) ([]*entity.ItemWithQuantity, error),
) error {
	return m.db.StartTransaction(func(tx *gorm.DB) (err error) {
		defer func() {
			if err != nil {
				logrus.Warnf("update stock transaction err=%v", err)
			}
		}()
		var dest []*persistent.StockModel
		if err = tx.Table("o_stock").
			Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).
			Where("product_id IN ?", getIDFromEntities(data)).Find(&dest).Error; err != nil {
			return errors.Wrap(err, "failed to find data")
		}
		existing := m.unmarshalFromDatabase(dest)

		//logrus.WithFields(logrus.Fields{
		//	"existing": existing,
		//	"data":     data,
		//}).Info("before update")

		updated, err := updateFn(ctx, existing, data)
		if err != nil {
			return err
		}

		//logrus.WithFields(logrus.Fields{
		//	"existing": existing,
		//	"data":     data,
		//	"updated":  updated,
		//}).Info("after update")

		for _, upd := range updated {
			if err = tx.Table("o_stock").Where("product_id = ?", upd.ID).Update("quantity", upd.Quantity).Error; err != nil {
				return errors.Wrapf(err, "failed to update %s", upd.ID)
			}
		}
		return nil
	})
}

func (m MySQLStockRepository) unmarshalFromDatabase(dest []*persistent.StockModel) []*entity.ItemWithQuantity {
	var result []*entity.ItemWithQuantity
	for _, i := range dest {
		result = append(result, &entity.ItemWithQuantity{
			ID:       i.ProductID,
			Quantity: i.Quantity,
		})
	}
	return result
}

func getIDFromEntities(items []*entity.ItemWithQuantity) []string {
	var ids []string
	for _, i := range items {
		ids = append(ids, i.ID)
	}
	return ids
}

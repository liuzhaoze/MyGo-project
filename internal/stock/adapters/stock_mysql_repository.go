package adapters

import (
	"context"
	"github.com/liuzhaoze/MyGo-project/stock/entity"
	"github.com/liuzhaoze/MyGo-project/stock/infrastructure/persistent"
	"github.com/liuzhaoze/MyGo-project/stock/infrastructure/persistent/builder"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
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
	query := builder.NewStock().ProductIDs(ids...)
	data, err := m.db.BatchGetStockByID(ctx, query)
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
		err = m.updatePessimistic(ctx, tx, data, updateFn)
		//err = m.updateOptimistic(ctx, tx, data, updateFn)
		return err
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

func (m MySQLStockRepository) updatePessimistic(
	ctx context.Context,
	tx *gorm.DB,
	data []*entity.ItemWithQuantity,
	updateFn func(ctx context.Context, existing []*entity.ItemWithQuantity, query []*entity.ItemWithQuantity) ([]*entity.ItemWithQuantity, error),
) error {
	var dest []*persistent.StockModel
	if err := builder.NewStock().ProductIDs(getIDFromEntities(data)...).ForUpdate().
		Fill(tx.Model(&persistent.StockModel{})).Find(&dest).Error; err != nil {
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
		for _, query := range data {
			if upd.ID == query.ID {
				if err = builder.NewStock().ProductIDs(upd.ID).QuantityGE(query.Quantity).Fill(tx.Model(&persistent.StockModel{})).
					Update("quantity", gorm.Expr("quantity - ?", query.Quantity)).Error; err != nil {
					return errors.Wrapf(err, "failed to update %s", upd.ID)
				}
			}
		}
	}
	return nil
}

func (m MySQLStockRepository) updateOptimistic(
	ctx context.Context,
	tx *gorm.DB,
	data []*entity.ItemWithQuantity,
	updateFn func(ctx context.Context, existing []*entity.ItemWithQuantity, query []*entity.ItemWithQuantity) ([]*entity.ItemWithQuantity, error),
) error {
	var dest []*persistent.StockModel
	if err := builder.NewStock().ProductIDs(getIDFromEntities(data)...).Fill(tx.Model(&persistent.StockModel{})).Find(&dest).Error; err != nil {
		return errors.Wrap(err, "failed to find data")
	}

	//logrus.WithFields(logrus.Fields{
	//	"existing": existing,
	//	"data":     data,
	//}).Info("before update")

	for _, queryData := range data {
		var latestRecord persistent.StockModel
		if err := builder.NewStock().ProductIDs(queryData.ID).Fill(tx.Model(&persistent.StockModel{})).First(&latestRecord).Error; err != nil {
			return err
		}

		if err := builder.NewStock().ProductIDs(queryData.ID).Versions(latestRecord.Version).QuantityGE(queryData.Quantity).Fill(tx.Model(&persistent.StockModel{})).
			Updates(map[string]any{"quantity": gorm.Expr("quantity - ?", queryData.Quantity), "version": latestRecord.Version + 1}).Error; err != nil {
			return err
		}
	}

	//logrus.WithFields(logrus.Fields{
	//	"existing": existing,
	//	"data":     data,
	//	"updated":  updated,
	//}).Info("after update")

	return nil
}

func getIDFromEntities(items []*entity.ItemWithQuantity) []string {
	var ids []string
	for _, i := range items {
		ids = append(ids, i.ID)
	}
	return ids
}

package persistent

import (
	"context"
	"fmt"
	"github.com/liuzhaoze/MyGo-project/common/logging"
	"github.com/liuzhaoze/MyGo-project/stock/infrastructure/persistent/builder"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type MySQL struct {
	db *gorm.DB
}

func NewMySQL() *MySQL {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		viper.GetString("mysql.user"),
		viper.GetString("mysql.password"),
		viper.GetString("mysql.host"),
		viper.GetString("mysql.port"),
		viper.GetString("mysql.dbname"))
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	logrus.Infof("dsn=%s", dsn)
	if err != nil {
		logrus.Panicf("connect to mysql failed, err=%v", err)
	}
	return &MySQL{db: db}
}

func NewMySQLWithDB(db *gorm.DB) *MySQL {
	return &MySQL{db: db}
}

func (m MySQL) UseTransaction(tx *gorm.DB) *gorm.DB {
	// Get 不使用事务，Create Update 使用事务
	if tx == nil {
		return m.db
	}
	return tx
}

func (m MySQL) Create(ctx context.Context, tx *gorm.DB, create *StockModel) (err error) {
	var returning StockModel
	_, deferLog := logging.WhenMySQL(ctx, "Create", create)
	defer deferLog(returning, &err)
	return m.UseTransaction(tx).WithContext(ctx).Model(&returning).Clauses(clause.Returning{}).Create(create).Error
}

func (m MySQL) GetStockByID(ctx context.Context, query *builder.Stock) (result *StockModel, err error) {
	_, deferLog := logging.WhenMySQL(ctx, "GetStockByID", query)
	defer deferLog(result, &err)

	err = query.Fill(m.db.WithContext(ctx)).First(&result).Error
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (m MySQL) BatchGetStockByID(ctx context.Context, query *builder.Stock) (result []StockModel, err error) {
	_, deferLog := logging.WhenMySQL(ctx, "BatchGetStockByID", query)
	defer deferLog(result, &err)

	err = query.Fill(m.db.WithContext(ctx)).Find(&result).Error
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (m MySQL) Update(ctx context.Context, tx *gorm.DB, cond *builder.Stock, update map[string]any) (err error) {
	var returning StockModel
	_, deferLog := logging.WhenMySQL(ctx, "UpdateStock", cond)
	defer deferLog(returning, &err)

	result := cond.Fill(m.UseTransaction(tx).WithContext(ctx).Model(&returning).Clauses(clause.Returning{})).Updates(update)
	return result.Error
}

func (m MySQL) StartTransaction(f func(tx *gorm.DB) error) error {
	return m.db.Transaction(f)
}

type StockModel struct {
	ID        int64     `gorm:"column:id"`
	ProductID string    `gorm:"column:product_id"`
	Quantity  int32     `gorm:"column:quantity"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
	Version   int64     `gorm:"column:version"`
}

func (m *StockModel) TableName() string {
	return "o_stock"
}

func (m *StockModel) BeforeCreate(tx *gorm.DB) (err error) {
	m.UpdatedAt = time.Now()
	return
}

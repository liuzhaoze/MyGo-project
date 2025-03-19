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

func (m MySQL) Create(ctx context.Context, create *StockModel) error {
	_, deferLog := logging.WhenMySQL(ctx, "Create", create)
	var returning StockModel
	err := m.db.WithContext(ctx).Model(&returning).Clauses(clause.Returning{}).Create(create).Error
	defer deferLog(returning, &err)
	return err
}

func (m MySQL) BatchGetStockByID(ctx context.Context, query *builder.Stock) ([]StockModel, error) {
	_, deferLog := logging.WhenMySQL(ctx, "BatchGetStockByID", query)
	var result []StockModel
	tx := query.Fill(m.db.WithContext(ctx).Clauses(clause.Returning{})).Find(&result)
	defer deferLog(result, &tx.Error)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return result, nil
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

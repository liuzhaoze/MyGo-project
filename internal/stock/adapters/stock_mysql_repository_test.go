package adapters

import (
	"context"
	"fmt"
	_ "github.com/liuzhaoze/MyGo-project/common/config"
	"github.com/liuzhaoze/MyGo-project/stock/entity"
	"github.com/liuzhaoze/MyGo-project/stock/infrastructure/persistent"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"sync"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) *persistent.MySQL {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		viper.GetString("mysql.user"),
		viper.GetString("mysql.password"),
		viper.GetString("mysql.host"),
		viper.GetString("mysql.port"),
		"",
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	assert.NoError(t, err)

	testDB := viper.GetString("mysql.dbname") + "_shadow"
	assert.NoError(t, db.Exec("DROP DATABASE IF EXISTS "+testDB).Error)
	assert.NoError(t, db.Exec("CREATE DATABASE IF NOT EXISTS "+testDB).Error)

	dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		viper.GetString("mysql.user"),
		viper.GetString("mysql.password"),
		viper.GetString("mysql.host"),
		viper.GetString("mysql.port"),
		testDB,
	)
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	assert.NoError(t, err)
	assert.NoError(t, db.AutoMigrate(persistent.StockModel{}))

	return persistent.NewMySQLWithDB(db)
}

func TestMySQLStockRepository_UpdateStock_race(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)

	var (
		ctx          = context.Background()
		testItem     = "test-race-item"
		initialStock = 100
	)

	err := db.Create(ctx, &persistent.StockModel{
		ProductID: testItem,
		Quantity:  int32(initialStock),
	})
	assert.NoError(t, err)

	repo := NewMySQLStockRepository(db)
	var wg sync.WaitGroup
	goroutines := 10
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		//time.Sleep(time.Duration(i) * time.Millisecond)
		//time.Sleep(200 * time.Millisecond)
		go func() {
			defer wg.Done()
			err := repo.UpdateStock(ctx, []*entity.ItemWithQuantity{
				{ID: testItem, Quantity: 1},
			},
				func(ctx context.Context, existing []*entity.ItemWithQuantity, query []*entity.ItemWithQuantity) ([]*entity.ItemWithQuantity, error) {
					var newItems []*entity.ItemWithQuantity
					for _, e := range existing {
						for _, q := range query {
							if e.ID == q.ID {
								newItems = append(newItems, &entity.ItemWithQuantity{
									ID:       e.ID,
									Quantity: e.Quantity - q.Quantity,
								})
							}
						}
					}
					return newItems, nil
				},
			)
			assert.NoError(t, err)
		}()
	}

	wg.Wait()

	res, err := db.BatchGetStockByID(ctx, []string{testItem})
	assert.NoError(t, err)
	assert.NotEmpty(t, res, "res cannot be empty")

	expected := initialStock - goroutines // 并发扣除 goroutines 个数量
	assert.EqualValues(t, expected, res[0].Quantity)
}

func TestMySQLStockRepository_UpdateStock_oversell(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)

	var (
		ctx          = context.Background()
		testItem     = "test-oversell-item"
		initialStock = 5
	)

	err := db.Create(ctx, &persistent.StockModel{
		ProductID: testItem,
		Quantity:  int32(initialStock),
	})
	assert.NoError(t, err)

	repo := NewMySQLStockRepository(db)
	var wg sync.WaitGroup
	goroutines := 100
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		//time.Sleep(time.Duration(i) * time.Millisecond)
		//time.Sleep(200 * time.Millisecond)
		go func() {
			defer wg.Done()
			err := repo.UpdateStock(ctx, []*entity.ItemWithQuantity{
				{ID: testItem, Quantity: 1},
			},
				func(ctx context.Context, existing []*entity.ItemWithQuantity, query []*entity.ItemWithQuantity) ([]*entity.ItemWithQuantity, error) {
					var newItems []*entity.ItemWithQuantity
					for _, e := range existing {
						for _, q := range query {
							if e.ID == q.ID {
								newItems = append(newItems, &entity.ItemWithQuantity{
									ID:       e.ID,
									Quantity: e.Quantity - q.Quantity,
								})
							}
						}
					}
					return newItems, nil
				},
			)
			assert.NoError(t, err)
		}()
		time.Sleep(10 * time.Millisecond)
	}

	wg.Wait()

	res, err := db.BatchGetStockByID(ctx, []string{testItem})
	assert.NoError(t, err)
	assert.NotEmpty(t, res, "res cannot be empty")

	// 5 个商品给 50 个 goroutines 买，最后剩余 0 个
	assert.GreaterOrEqual(t, res[0].Quantity, int32(0))
}

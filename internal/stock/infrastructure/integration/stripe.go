package integration

import (
	"context"
	_ "github.com/liuzhaoze/MyGo-project/common/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/product"
)

type StripeAPI struct {
	apiKey string
}

func NewStripeAPI() *StripeAPI {
	key := viper.GetString("stripe-key")
	if key == "" {
		logrus.Fatal("empty stripe-key")
	}
	return &StripeAPI{apiKey: viper.GetString("stripe-key")}
}

func (s *StripeAPI) GetPriceByProductID(ctx context.Context, prodID string) (string, error) {
	stripe.Key = s.apiKey

	result, err := product.Get(prodID, &stripe.ProductParams{})
	if err != nil {
		return "", err
	}
	return result.DefaultPrice.ID, err
}

func (s *StripeAPI) GetProductByID(ctx context.Context, prodID string) (*stripe.Product, error) {
	stripe.Key = s.apiKey
	return product.Get(prodID, &stripe.ProductParams{})
}

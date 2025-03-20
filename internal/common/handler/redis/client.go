package redis

import (
	"context"
	"errors"
	"github.com/liuzhaoze/MyGo-project/common/logging"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"time"
)

func SetNX(ctx context.Context, client *redis.Client, key, value string, ttl time.Duration) (err error) {
	now := time.Now()
	defer func() {
		l := logrus.WithContext(ctx).WithFields(logrus.Fields{
			"start":       now,
			"key":         key,
			"value":       value,
			logging.Error: err,
			logging.Cost:  time.Since(now).Milliseconds(),
		})

		if err == nil {
			l.Info("_redis_setnx_success")
		} else {
			l.Error("_redis_setnx_fail")
		}
	}()

	if client == nil {
		return errors.New("redis client is nil")
	}
	_, err = client.SetNX(ctx, key, value, ttl).Result()
	return err
}
func Del(ctx context.Context, client *redis.Client, key string) (err error) {
	now := time.Now()
	defer func() {
		l := logrus.WithContext(ctx).WithFields(logrus.Fields{
			"start":       now,
			"key":         key,
			logging.Error: err,
			logging.Cost:  time.Since(now).Milliseconds(),
		})

		if err == nil {
			l.Info("_redis_del_success")
		} else {
			l.Error("_redis_del_fail")
		}
	}()

	if client == nil {
		return errors.New("redis client is nil")
	}
	_, err = client.Del(ctx, key).Result()
	return err
}

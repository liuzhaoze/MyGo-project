package redis

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"time"
)

func SetNX(ctx context.Context, client *redis.Client, key, value string, ttl time.Duration) (err error) {
	now := time.Now()
	defer func() {
		l := logrus.WithContext(ctx).WithFields(logrus.Fields{
			"start": now,
			"key":   key,
			"value": value,
			"error": err,
			"cost":  time.Since(now).Milliseconds(),
		})

		if err == nil {
			l.Info("redis_setnx success")
		} else {
			l.Error("redis_setnx fail")
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
			"start": now,
			"key":   key,
			"error": err,
			"cost":  time.Since(now).Milliseconds(),
		})

		if err == nil {
			l.Info("redis_del success")
		} else {
			l.Error("redis_del fail")
		}
	}()

	if client == nil {
		return errors.New("redis client is nil")
	}
	_, err = client.Del(ctx, key).Result()
	return err
}

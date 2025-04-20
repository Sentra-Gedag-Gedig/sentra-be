package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"time"
)

type IRedis interface {
	SetOTP(ctx context.Context, key string, code string, expiration time.Duration) error
	GetOTP(ctx context.Context, key string) (string, error)
}

type redisClient struct {
	client *redis.Client
}

func New() IRedis {
	db, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	redisAddr := os.Getenv("REDIS_ADDRESS")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	logrus.Info(fmt.Sprintf("Connecting to Redis at %s...", redisAddr))

	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		logrus.Error(fmt.Sprintf("Failed to connect to Redis: %v", err))
	} else {
		logrus.Info("Successfully connected to Redis")
	}

	return &redisClient{client: client}
}

func (r *redisClient) SetOTP(ctx context.Context, key string, code string, expiration time.Duration) error {
	logrus.Debug(fmt.Sprintf("Setting OTP for key %s with expiration %v", key, expiration))
	err := r.client.Set(ctx, key, code, expiration).Err()
	if err != nil {
		logrus.Error(fmt.Sprintf("Error setting OTP for key %s: %v", key, err))
		return err
	}
	logrus.Debug(fmt.Sprintf("Successfully set OTP for key %s", key))
	return nil
}

func (r *redisClient) GetOTP(ctx context.Context, key string) (string, error) {
	logrus.Debug(fmt.Sprintf("Getting OTP for key %s", key))
	val, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		logrus.Debug(fmt.Sprintf("OTP not found for key %s", key))
		return "", err
	} else if err != nil {
		logrus.Error(fmt.Sprintf("Error getting OTP for key %s: %v", key, err))
		return "", err
	}
	logrus.Debug(fmt.Sprintf("Successfully got OTP for key %s", key))
	return val, nil
}

func (r *redisClient) DeleteOTP(ctx context.Context, key string) error {
	logrus.Debug(fmt.Sprintf("Deleting OTP for key %s", key))
	result, err := r.client.Del(ctx, key).Result()
	if err != nil {
		logrus.Error(fmt.Sprintf("Error deleting OTP for key %s: %v", key, err))
		return err
	}

	if result == 0 {
		logrus.Debug(fmt.Sprintf("OTP key %s not found for deletion", key))
		return nil
	}

	logrus.Debug(fmt.Sprintf("Successfully deleted OTP for key %s", key))
	return nil
}

package kvclient

import (
	"context"
	"time"
	"travel_ai_search/search/conf"

	"github.com/redis/go-redis/v9"
)

type KVClient struct {
	cli *redis.Client
	ctx context.Context
	//timeout time.Duration
}

var kvCli *KVClient

func GetInstance() *KVClient {
	return kvCli
}
func InitClient(config *conf.Config) (*KVClient, error) {
	var client KVClient
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       0,
	})
	client.cli = redisClient
	client.ctx = context.Background()
	kvCli = &client
	return &client, nil
}

func (client *KVClient) Close() {
	client.cli.Close()
}

func (client *KVClient) Set(key string, value interface{}, expired time.Duration) error {
	_, err := client.cli.Set(client.ctx, key, value, expired).Result()
	return err
}

func (client *KVClient) Get(key string) (string, error) {
	val, err := client.cli.Get(client.ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

func (client *KVClient) IncrBy(key string, step int64) (int64, error) {
	val, err := client.cli.IncrBy(client.ctx, key, step).Result()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

// func (client *KVClient) HSet(key string, values ...interface{}) error {
// 	_, err := client.cli.HSet(client.ctx, key, values...).Result()
// 	return err
// }

func (client *KVClient) HMSet(key string, values ...interface{}) error {
	_, err := client.cli.HMSet(client.ctx, key, values...).Result()
	return err
}

func (client *KVClient) HGetAll(key string) (map[string]string, error) {
	val, err := client.cli.HGetAll(client.ctx, key).Result()
	if err == redis.Nil {
		return make(map[string]string), nil
	}
	return val, err
}

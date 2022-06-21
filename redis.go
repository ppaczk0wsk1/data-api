package main

import (
	"github.com/go-redis/redis"
)

func newRedisClient() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password
		DB:       0,  // default DB
	})
	return rdb
}

package AppConfig

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
)

func ConnectRedis() (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB: func() int {
			db, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
			return db
		}(),
	})

	_, err := client.Ping(client.Context()).Result()
	if err != nil {
		return nil, fmt.Errorf("error connecting to redis")
	}
	log.Println("Connected to redis")
	return client, nil
}
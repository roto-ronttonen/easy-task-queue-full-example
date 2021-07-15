package main

import (
	"context"
	"math/rand"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	easytaskqueueclientgo "github.com/roto-ronttonen/easy-task-queue-client-go"
)

func add(redisClient *redis.Client, ctx context.Context) func() {
	return func() {
		time.Sleep(15 * time.Second)
		redisClient.Publish(ctx, "taskdone", rand.Int())
	}
}

func main() {
	taskQueueClient := easytaskqueueclientgo.NewWorkerClient(os.Getenv("TASK_QUEUE_ADDRESS"), "add")
	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDRESS"),
		Password: "",
		DB:       0,
	})
	ctx := context.Background()

	taskQueueClient.Start(add(redisClient, ctx))
}

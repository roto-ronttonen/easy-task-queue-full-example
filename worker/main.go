package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	easytaskqueueclientgo "github.com/roto-ronttonen/easy-task-queue-client-go"
)

func add(redisClient *redis.Client, ctx context.Context) func(data string) {
	return func(data string) {
		time.Sleep(7 * time.Second)
		if len(data) == 0 {
			redisClient.Publish(ctx, "taskdone", rand.Int())
		} else {
			redisClient.Publish(ctx, "taskdone", data)
		}

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

	err := taskQueueClient.Start(add(redisClient, ctx))

	log.Fatal(err.Error())
}

package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-redis/redis/v8"
	easytaskqueueclientgo "github.com/roto-ronttonen/easy-task-queue-client-go"
)

type RequestState struct {
	responseChannel chan []int
}

func stateHandler(stateRequestChan chan RequestState, addDataChan chan int) {
	var data []int

	for {
		select {
		case req := <-stateRequestChan:
			req.responseChannel <- data
		case add := <-addDataChan:
			data = append(data, add)
		}

	}
}

func pubSubHandler(redisClient *redis.Client, addDataChan chan int) {
	ctx := context.Background()
	pubsub := redisClient.Subscribe(ctx, "taskdone")
	channel := pubsub.Channel()

	for msg := range channel {
		data, err := strconv.Atoi(string(msg.Payload))
		if err != nil {
			log.Print(err.Error())
		} else {
			addDataChan <- data
		}
	}

}

func main() {
	stateRequestChan := make(chan RequestState)
	addDataChan := make(chan int)
	taskQueueClient := easytaskqueueclientgo.NewClient(os.Getenv("TASK_QUEUE_ADDRESS"))
	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDRESS"),
		Password: "",
		DB:       0,
	})

	go stateHandler(stateRequestChan, addDataChan)
	go pubSubHandler(redisClient, addDataChan)
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/api/data", func(w http.ResponseWriter, r *http.Request) {
		resChan := make(chan []int)
		stateRequestChan <- RequestState{
			responseChannel: resChan,
		}
		data := <-resChan
		jsonData, err := json.Marshal(data)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(jsonData))
	})
	r.Post("/api/data", func(w http.ResponseWriter, r *http.Request) {
		err := taskQueueClient.SendTask("add")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		resChan := make(chan []int)
		stateRequestChan <- RequestState{
			responseChannel: resChan,
		}
		data := <-resChan
		jsonData, err := json.Marshal(data)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(jsonData))
	})
	http.ListenAndServe("0.0.0.0:3000", r)

}

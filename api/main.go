package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-redis/redis/v8"
	easytaskqueueclientgo "github.com/roto-ronttonen/easy-task-queue-client-go"
)

type RequestState struct {
	responseChannel chan []string
}

func stateHandler(stateRequestChan chan RequestState, addDataChan chan string) {
	var data []string

	for {
		select {
		case req := <-stateRequestChan:
			req.responseChannel <- data
		case add := <-addDataChan:
			data = append(data, add)
		}

	}
}

func pubSubHandler(redisClient *redis.Client, addDataChan chan string) {
	ctx := context.Background()
	pubsub := redisClient.Subscribe(ctx, "taskdone")
	channel := pubsub.Channel()

	for msg := range channel {
		data := string(msg.Payload)

		addDataChan <- data

	}

}

type postBody struct {
	OptionalMessage string `json:"optionalMessage"`
}

func main() {
	stateRequestChan := make(chan RequestState)
	addDataChan := make(chan string)
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
		resChan := make(chan []string)
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
		var p postBody

		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if len(p.OptionalMessage) == 0 {
			err = taskQueueClient.SendTask("add")
		} else {
			err = taskQueueClient.SendTaskWithData("add", p.OptionalMessage)
		}

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		resChan := make(chan []string)
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

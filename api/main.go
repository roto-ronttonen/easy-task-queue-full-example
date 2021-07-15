package main

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

func main() {

	taskQueueClient := easytaskqueueclientgo.NewClient(os.Getenv("TASK_QUEUE_ADDRESS"))

	stateRequestChan := make(chan RequestState)
	addDataChan := make(chan int)

	go stateHandler(stateRequestChan, addDataChan)

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
	http.ListenAndServe(":3000", r)

}

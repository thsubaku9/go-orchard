package worker

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"orchard/task"

	"github.com/gorilla/mux"
)

type HttpApi struct {
	Address string
	Port    int
	Worker  *Worker
	Router  *mux.Router
}

type StandardResponse struct {
	HttpStatusCode int
	ErrorMsg       string
	Response       interface{}
}

func (httpApi *HttpApi) StartTaskHandler(w http.ResponseWriter, r *http.Request) {
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()

	ts := task.TaskEvent{}
	err := d.Decode(ts)

	if err != nil {
		msg := fmt.Sprintf("Error unmarshalling body: %v\n", err)
		log.Print(msg)
		w.WriteHeader(400)
		e := StandardResponse{
			HttpStatusCode: 400,
			ErrorMsg:       msg,
		}
		json.NewEncoder(w).Encode(e)
		return
	}

	httpApi.Worker.AddTask(ts.Task)
	log.Printf("Added task %v\n", ts.Task.ID)
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(ts.Task)

}

func (httpApi *HttpApi) StopTaskHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

}

func (httpApi *HttpApi) ListAllTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(httpApi.Worker.ListTasks())

}

func (httpApi *HttpApi) GetTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(httpApi.Worker.GetTask(vars["containerId"]))
}

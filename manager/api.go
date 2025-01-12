package manager

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"orchard/api"
	"orchard/task"
	"time"

	"github.com/gorilla/mux"
)

type HttpApiManager struct {
	api.HttpApi[Manager]
}

func (a *HttpApiManager) StartTaskHandler(w http.ResponseWriter, r *http.Request) {

	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	te := task.TaskEvent{}
	err := d.Decode(&te)
	if err != nil {
		msg := fmt.Sprintf("Error unmarshalling body: %v\n", err)
		log.Printf(msg)
		w.WriteHeader(http.StatusNotFound)
		e := api.StandardResponse[any]{
			HttpStatusCode: http.StatusNotFound,
			ErrorMsg:       msg,
		}
		json.NewEncoder(w).Encode(e)
		return
	}
	a.Ref.AddTask(te)
	log.Printf("Added task %v\n", te.Task.ID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(api.StandardResponse[task.Task]{
		HttpStatusCode: http.StatusCreated,
		Response:       te.Task,
	})

}

func (a *HttpApiManager) GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(a.Ref.GetTasks())

}

func (a *HttpApiManager) StopTaskHandler(w http.ResponseWriter, r *http.Request) {

}

func (httpApi *HttpApiManager) initRouter() {
	httpApi.Router = mux.NewRouter()

	httpApi.Router.HandleFunc("/tasks", httpApi.GetTasksHandler).Methods("GET")
	httpApi.Router.HandleFunc("/tasks", httpApi.StartTaskHandler).Methods("POST")
	httpApi.Router.HandleFunc("/tasks/{taskId}", httpApi.StopTaskHandler).Methods("DELETE")
}

func (httpApi *HttpApiManager) StartServer() {

	httpApi.initRouter()
	server := http.Server{
		Handler:      httpApi.Router,
		Addr:         fmt.Sprintf("%s:%s", httpApi.Address, httpApi.Port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("Hosting on %s:%s\n", httpApi.Address, httpApi.Port)
	api.PrintEndpoints(httpApi.Router)
	server.ListenAndServe()
}

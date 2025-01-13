package manager

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"orchard/api"
	"orchard/task"
	"time"

	"github.com/google/uuid"
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
	vars := mux.Vars(r)
	taskId := vars["taskId"]

	if taskId == "" {
		w.WriteHeader(http.StatusNotAcceptable)

		json.NewEncoder(w).Encode(api.StandardResponse[any]{
			HttpStatusCode: http.StatusNotAcceptable,
			ErrorMsg:       "Empty taskId passed",
		})
		return
	}

	tID, _ := uuid.Parse(taskId)
	taskToStop, ok := a.Ref.TaskDb[tID]

	if !ok {
		w.WriteHeader(http.StatusNotFound)

		json.NewEncoder(w).Encode(api.StandardResponse[any]{
			HttpStatusCode: http.StatusNotFound,
			ErrorMsg:       "Task not found",
		})
		return
	}

	te := task.TaskEvent{ID: uuid.New(),
		State:     task.Completed,
		Timestamp: time.Now(),
	}
	taskCopy := *taskToStop
	taskCopy.State = task.Completed
	te.Task = taskCopy
	a.Ref.AddTask(te)
	log.Printf("Added task event %v to stop task %v\n", te.ID, taskToStop.ID)
	w.WriteHeader(http.StatusNoContent)

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

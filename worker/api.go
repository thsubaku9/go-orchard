package worker

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"orchard/api"
	"orchard/metrics"
	"orchard/task"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type HttpApiWorker struct {
	api.HttpApi[Worker]
}

func (httpApiWorker *HttpApiWorker) StartTaskHandler(w http.ResponseWriter, r *http.Request) {
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()

	ts := task.TaskEvent{}
	err := d.Decode(&ts)

	if err != nil {
		msg := fmt.Sprintf("Error unmarshalling body: %v\n", err)
		log.Print(msg)
		w.WriteHeader(http.StatusNotFound)
		e := api.StandardResponse[any]{
			HttpStatusCode: http.StatusNotFound,
			ErrorMsg:       msg,
		}
		json.NewEncoder(w).Encode(e)
		return
	}

	ts.Task.State = task.Pending
	ts.Task.Event = task.SpinUp

	httpApiWorker.Ref.AddTask(ts.Task)
	log.Printf("Added task %v\n", ts.Task.ID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(api.StandardResponse[task.Task]{
		HttpStatusCode: http.StatusCreated,
		Response:       ts.Task,
	})

}

func (httpApiWorker *HttpApiWorker) StopTaskHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskId := vars["taskId"]

	if taskId == "" {
		log.Printf("No taskID passed in request.\n")
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(api.StandardResponse[any]{
			HttpStatusCode: http.StatusBadRequest,
			ErrorMsg:       "Empty taskId passed",
		})
		return
	}

	tID, _ := uuid.Parse(taskId)
	taskToStop, ok := httpApiWorker.Ref.Db[tID]

	if !ok {
		log.Printf("No task with ID %v found", tID)
		w.WriteHeader(http.StatusNotFound)

		json.NewEncoder(w).Encode(api.StandardResponse[any]{
			HttpStatusCode: http.StatusNotFound,
			ErrorMsg:       fmt.Sprintf("No task with ID %v found", tID),
		})
		return
	}

	taskCopy := *taskToStop
	taskCopy.Event = task.SpinDown
	httpApiWorker.Ref.AddTask(taskCopy)

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(api.StandardResponse[task.Task]{
		HttpStatusCode: http.StatusOK,
		Response:       taskCopy,
	})

}

func (httpApiWorker *HttpApiWorker) ListAllTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(api.StandardResponse[[]task.Task]{
		HttpStatusCode: http.StatusOK,
		Response:       httpApiWorker.Ref.ListTasks(),
	})
}

func (httpApiWorker *HttpApiWorker) ListAllTaskIds(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(api.StandardResponse[[]uuid.UUID]{
		HttpStatusCode: http.StatusOK,
		Response:       httpApiWorker.Ref.ListTaskIds(),
	})
}

func (httpApiWorker *HttpApiWorker) GetTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tID, _ := uuid.Parse(vars["taskId"])

	if vars["taskId"] == "" {
		log.Printf("No taskID passed in request.\n")
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(api.StandardResponse[any]{
			HttpStatusCode: http.StatusBadRequest,
			ErrorMsg:       "Empty taskId passed",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(
		api.StandardResponse[task.DockerInspectResponse]{
			HttpStatusCode: http.StatusOK,
			Response:       httpApiWorker.Ref.GetTask(tID),
		})
}

func (httpApiWorker *HttpApiWorker) GetStatsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(api.StandardResponse[metrics.Metrics]{
		HttpStatusCode: http.StatusOK,
		Response:       metrics.GetFullMetrics(),
	})
}

func (httpApiWorker *HttpApiWorker) initRouter() {
	httpApiWorker.Router = mux.NewRouter()

	httpApiWorker.Router.HandleFunc("/tasks", httpApiWorker.ListAllTasks).Methods("GET")
	httpApiWorker.Router.HandleFunc("/tasks/ids", httpApiWorker.ListAllTasks).Methods("GET")
	httpApiWorker.Router.HandleFunc("/tasks/{taskId}", httpApiWorker.GetTask).Methods("GET")
	httpApiWorker.Router.HandleFunc("/tasks", httpApiWorker.StartTaskHandler).Methods("POST")
	httpApiWorker.Router.HandleFunc("/tasks/{taskId}", httpApiWorker.StopTaskHandler).Methods("DELETE")

	httpApiWorker.Router.HandleFunc("/stats", httpApiWorker.GetStatsHandler).Methods("GET")
}

func (httpApiWorker *HttpApiWorker) StartServer() {

	httpApiWorker.initRouter()
	server := http.Server{
		Handler:      httpApiWorker.Router,
		Addr:         fmt.Sprintf("%s:%s", httpApiWorker.Address, httpApiWorker.Port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("Hosting on %s:%s\n", httpApiWorker.Address, httpApiWorker.Port)
	api.PrintEndpoints(httpApiWorker.Router)
	server.ListenAndServe()

}

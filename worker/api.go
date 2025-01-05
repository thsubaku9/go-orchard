package worker

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"orchard/task"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type HttpApi struct {
	Address string
	Port    string
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
	err := d.Decode(&ts)

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

	ts.Task.State = task.Pending
	ts.Task.Event = task.SpinUp

	httpApi.Worker.AddTask(ts.Task)
	log.Printf("Added task %v\n", ts.Task.ID)
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(StandardResponse{
		HttpStatusCode: 201,
		Response:       ts.Task,
	})

}

func (httpApi *HttpApi) StopTaskHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskId := vars["taskId"]

	if taskId == "" {
		log.Printf("No taskID passed in request.\n")
		w.WriteHeader(400)

		json.NewEncoder(w).Encode(StandardResponse{
			HttpStatusCode: 201,
			ErrorMsg:       "Empty taskId passed",
		})
		return
	}

	tID, _ := uuid.Parse(taskId)
	taskToStop, ok := httpApi.Worker.Db[tID]

	if !ok {
		log.Printf("No task with ID %v found", tID)
		w.WriteHeader(404)

		json.NewEncoder(w).Encode(StandardResponse{
			HttpStatusCode: 404,
			ErrorMsg:       fmt.Sprintf("No task with ID %v found", tID),
		})
		return
	}

	taskCopy := *taskToStop
	taskCopy.Event = task.SpinDown
	httpApi.Worker.AddTask(taskCopy)

	w.WriteHeader(200)

	json.NewEncoder(w).Encode(StandardResponse{
		HttpStatusCode: 200,
		Response:       taskCopy,
	})

}

func (httpApi *HttpApi) ListAllTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(StandardResponse{
		HttpStatusCode: 200,
		Response:       httpApi.Worker.ListTasks(),
	})

}

func (httpApi *HttpApi) GetTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tID, _ := uuid.Parse(vars["taskId"])

	if vars["taskId"] == "" {
		log.Printf("No taskID passed in request.\n")
		w.WriteHeader(400)

		json.NewEncoder(w).Encode(StandardResponse{
			HttpStatusCode: 201,
			ErrorMsg:       "Empty taskId passed",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(httpApi.Worker.GetTask(tID))
}

func (httpApi *HttpApi) GetStatsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(StandardResponse{
		HttpStatusCode: 200,
		Response:       GetFullMetrics(),
	})
}

func (httpApi *HttpApi) initRouter() {
	httpApi.Router = mux.NewRouter()

	httpApi.Router.HandleFunc("/tasks", httpApi.ListAllTasks).Methods("GET")
	httpApi.Router.HandleFunc("/tasks/{taskId}", httpApi.GetTask).Methods("GET")
	httpApi.Router.HandleFunc("/tasks", httpApi.StartTaskHandler).Methods("POST")
	httpApi.Router.HandleFunc("/tasks/{taskId}", httpApi.StopTaskHandler).Methods("DELETE")

	httpApi.Router.HandleFunc("/stats", httpApi.GetStatsHandler).Methods("GET")
}

func (httpApi *HttpApi) StartServer() {

	httpApi.initRouter()
	server := http.Server{
		Handler:      httpApi.Router,
		Addr:         fmt.Sprintf("%s:%s", httpApi.Address, httpApi.Port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("Hosting on %s:%s\n", httpApi.Address, httpApi.Port)
	printEndpoints(httpApi.Router)
	server.ListenAndServe()

}

func printEndpoints(r *mux.Router) {
	r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, err := route.GetPathTemplate()
		if err != nil {
			return err
		}
		methods, err := route.GetMethods()
		if err != nil {
			return err
		}
		log.Printf("%v %s\n", methods, path)
		return nil
	})
}

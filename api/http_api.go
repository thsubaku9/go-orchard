package api

import (
	"log"

	"github.com/gorilla/mux"
)

type HttpApi[REF any] struct {
	Address string
	Port    string
	Ref     *REF
	Router  *mux.Router
}

func PrintEndpoints(r *mux.Router) {
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

type StandardResponse[R any] struct {
	HttpStatusCode int
	ErrorMsg       string
	Response       R
}

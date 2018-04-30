package router

import (
	"net/http"

	"github.com/gorilla/mux"
)



type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}


type ForumAPI map[string]Route

func CreateRouter(startpoint string, routes *ForumAPI) *mux.Router {

	newRouter := mux.NewRouter().StrictSlash(true)

	for _, route := range *routes {
		newRouter.
			Methods(route.Method).
			Path(startpoint + route.Pattern).
			Name(route.Name).
			Handler(route.HandlerFunc)
	}

	return newRouter
}

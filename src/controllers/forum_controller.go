package controllers

import (
	"net/http"
	"strconv"

	resp "github.com/Alex-Kuz/tp-database/src/utils/responses"
	"github.com/gorilla/mux"
)

// optimized?
func ForumDetails(respWriter http.ResponseWriter, request *http.Request) {
	respWriter.Header().Set("Content-Type", "application/json; charset=utf-8")

	slug := mux.Vars(request)["slug"]

	forum := ForumService.GetForumBySlug(slug)
	if forum == nil {
		respWriter.WriteHeader(http.StatusNotFound)
		writeJSONBody(&respWriter, resp.Message{"Forum not found"})
		return
	}

	respWriter.WriteHeader(http.StatusOK)
	writeJSONBody(&respWriter, *forum)
}

//  opt ?
func ForumThreads(respWriter http.ResponseWriter, request *http.Request) {
	respWriter.Header().Set("Content-Type", "application/json; charset=utf-8")

	slug := mux.Vars(request)["slug"]

	limit := request.URL.Query().Get("limit")
	since := request.URL.Query().Get("since")
	descRef := request.URL.Query().Get("desc")

	desc := false
	if descRef != "" {
		var err error
		desc, err = strconv.ParseBool(descRef)
		if err != nil {
			panic(err)
		}
	}

	forumExists, threads := ThreadService.SelectThreads(slug, limit, since, desc)

	if !forumExists {
		respWriter.WriteHeader(http.StatusNotFound)
		writeJSONBody(&respWriter, resp.Message{"Forum not found"})
		return
	}

	respWriter.WriteHeader(http.StatusOK)
	writeJSONBody(&respWriter, threads)
}

func ForumUsers(respWriter http.ResponseWriter, request *http.Request) {
	respWriter.Header().Set("Content-Type", "application/json; charset=utf-8")

	threadSlug := mux.Vars(request)["slug"]

	forum := ForumService.GetForumBySlug(threadSlug)

	if forum == nil {
		respWriter.WriteHeader(http.StatusNotFound)
		writeJSONBody(&respWriter, resp.Message{"forum not found"})
		return
	}

	limit := request.URL.Query().Get("limit")
	since := request.URL.Query().Get("since")
	descRef := request.URL.Query().Get("desc")

	desc := false
	if descRef != "" {
		var err error
		desc, err = strconv.ParseBool(descRef)
		if err != nil {
			panic(err)
		}
	}

	users := ForumService.GetUsers(forum, since, limit, desc)

	respWriter.WriteHeader(http.StatusOK)
	writeJSONBody(&respWriter, users)
}

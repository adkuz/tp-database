package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Alex-Kuz/tp-database/src/models"
	resp "github.com/Alex-Kuz/tp-database/src/utils/responses"
	"github.com/gorilla/mux"
)

func ThreadDetails(respWriter http.ResponseWriter, request *http.Request) {
	respWriter.Header().Set("Content-Type", "application/json; charset=utf-8")

	threadSlug := mux.Vars(request)["slug_or_id"]

	var thread *models.Thread
	threadID, err := strconv.ParseUint(threadSlug, 10, 64)
	if err == nil {
		thread = ThreadService.GetThreadById(threadID)
	} else {
		thread = ThreadService.GetThreadBySlug(threadSlug)
	}

	if thread == nil {
		respWriter.WriteHeader(http.StatusNotFound)
		writeJSONBody(&respWriter, resp.Message{"Forum not found"})
		return
	}

	respWriter.WriteHeader(http.StatusOK)
	writeJSONBody(&respWriter, *thread)
}

func ThreadPosts(respWriter http.ResponseWriter, request *http.Request) {
	respWriter.Header().Set("Content-Type", "application/json; charset=utf-8")

	threadSlug := mux.Vars(request)["slug_or_id"]

	var thread *models.Thread
	threadId, err := strconv.ParseUint(threadSlug, 10, 64)
	if err == nil {
		thread = ThreadService.GetThreadById(threadId)
	} else {
		thread = ThreadService.GetThreadBySlug(threadSlug)
	}

	if thread == nil {
		respWriter.WriteHeader(http.StatusNotFound)
		writeJSONBody(&respWriter, resp.Message{"Thread not found"})
		return
	}

	limit := request.URL.Query().Get("limit")
	since := request.URL.Query().Get("since")
	sort := request.URL.Query().Get("sort")

	descRef := request.URL.Query().Get("desc")
	desc := false
	if descRef != "" {
		var err error
		desc, err = strconv.ParseBool(descRef)
		if err != nil {
			panic(err)
		}
	}

	var posts []models.Post
	if sort == "parent_tree" {
		posts = PostService.GetPostsParentTreeSort(thread, limit, since, desc)
	} else if sort == "tree" {
		posts = PostService.GetPostsTreeSort(thread, limit, since, desc)
	} else {
		posts = PostService.GetPostsFlat(thread, limit, since, desc)
	}

	respWriter.WriteHeader(http.StatusOK)
	writeJSONBody(&respWriter, posts)
}

// does not perf ----------------------------------------------------------------------

func ThreadUpdate(respWriter http.ResponseWriter, request *http.Request) {
	respWriter.Header().Set("Content-Type", "application/json; charset=utf-8")

	threadSlug := mux.Vars(request)["slug_or_id"]

	var thread *models.Thread
	threadId, err := strconv.ParseUint(threadSlug, 10, 64)
	if err == nil {
		thread = ThreadService.GetThreadById(threadId)
	} else {
		thread = ThreadService.GetThreadBySlug(threadSlug)
	}

	if thread == nil {
		respWriter.WriteHeader(http.StatusNotFound)
		writeJSONBody(&respWriter, resp.Message{"Forum not found"})
		return
	}

	var threadMap map[string]string

	if err := json.NewDecoder(request.Body).Decode(&threadMap); err != nil {
		panic(err)
	}

	if value, ok := threadMap["message"]; ok {
		thread.Message = value
	}

	if value, ok := threadMap["title"]; ok {
		thread.Title = value
	}

	ThreadService.UpdateThread(thread)

	respWriter.WriteHeader(http.StatusOK)
	writeJSONBody(&respWriter, *thread)
}

package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	resp "github.com/Alex-Kuz/tp-database/src/utils/responses"
	"github.com/gorilla/mux"
)

func PostDetails(respWriter http.ResponseWriter, request *http.Request) {
	respWriter.Header().Set("Content-Type", "application/json; charset=utf-8")

	id, err := strconv.ParseUint(mux.Vars(request)["id"], 10, 64)
	if err != nil {
		panic(err)
	}

	post := PostService.GetPostById(id)

	if post == nil {
		respWriter.WriteHeader(http.StatusNotFound)
		writeJSONBody(&respWriter, resp.Message{"Post not found"})
		return
	}

	var postInfo resp.PostInfo
	postInfo.Post = post

	var list []string
	if related := request.URL.Query()["related"]; len(related) > 0 {
		list = strings.Split(related[0], ",")
	}

	for i := 0; i < len(list); i++ {
		// fmt.Println(list[i])
		if list[i] == "user" {
			// fmt.Println(list, "USER")
			if user := UserService.GetUserByNickname(post.Author); user != nil {
				postInfo.Author = user
			}
		} else if list[i] == "thread" {
			// fmt.Println(list, "THREAD")
			if thread := ThreadService.GetThreadById(post.Thread); thread != nil {
				postInfo.Thread = thread
			}
		} else if list[i] == "forum" {
			// fmt.Println(list, "FORUM")
			if forum := ForumService.GetForumBySlug(post.Forum); forum != nil {
				postInfo.Forum = forum
			}
		}
	}

	respWriter.WriteHeader(http.StatusOK)
	writeJSONBody(&respWriter, postInfo)
}

// does not perf
func PostUpdate(respWriter http.ResponseWriter, request *http.Request) {
	respWriter.Header().Set("Content-Type", "application/json; charset=utf-8")

	id, err := strconv.ParseUint(mux.Vars(request)["id"], 10, 64)
	if err != nil {
		panic(err)
	}

	post := PostService.GetPostById(id)

	if post == nil {
		respWriter.WriteHeader(http.StatusNotFound)
		writeJSONBody(&respWriter, resp.Message{"Post not found"})
		return
	}

	var updateMap map[string]string
	if err := json.NewDecoder(request.Body).Decode(&updateMap); err != nil {
		panic(err)
	}

	if value, ok := updateMap["message"]; ok && value != "" && value != post.Message {
		post.Message = value
		PostService.UpdatePost(post)
	}

	respWriter.WriteHeader(http.StatusOK)
	writeJSONBody(&respWriter, post)
}

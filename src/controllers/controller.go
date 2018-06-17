package controllers

import (
	"encoding/json"
	_ "fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/joeshaw/iso8601"

	"github.com/Alex-Kuz/tp-database/src/models"
	resp "github.com/Alex-Kuz/tp-database/src/utils/responses"
)

// optimized with index
func UserProfile(respWriter http.ResponseWriter, request *http.Request) {
	respWriter.Header().Set("Content-Type", "application/json; charset=utf-8")

	nickname := mux.Vars(request)["nickname"]

	user := UserService.GetUserByNickname(nickname)

	if user != nil {
		respWriter.WriteHeader(http.StatusOK)
		writeJSONBody(&respWriter, *user)
	} else {
		respWriter.WriteHeader(http.StatusNotFound)
		writeJSONBody(&respWriter, resp.MsgCantFindUser(nickname))
	}
}

func UpdateUser(respWriter http.ResponseWriter, request *http.Request) {
	respWriter.Header().Set("Content-Type", "application/json; charset=utf-8")

	nickname := mux.Vars(request)["nickname"]

	user := UserService.GetUserByNickname(nickname)
	if user == nil {
		respWriter.WriteHeader(http.StatusNotFound)
		writeJSONBody(&respWriter, resp.MsgCantFindUser(nickname))
		return
	}

	var userMap map[string]string
	var userInfo models.User

	json.NewDecoder(request.Body).Decode(&userMap)

	if value, ok := userMap["email"]; ok {
		userInfo.Email = value
	} else {
		userInfo.Email = user.Email
	}

	if value, ok := userMap["about"]; ok {
		userInfo.About = value
	} else {
		userInfo.About = user.About
	}

	if value, ok := userMap["fullname"]; ok {
		userInfo.Fullname = value
	} else {
		userInfo.Fullname = user.Fullname
	}

	userInfo.Nickname = nickname

	// Конфликт может возникнуть только по значению email
	if opponent := UserService.GetUserByEmail(userInfo.Email); opponent != nil && *opponent != *user {
		respWriter.WriteHeader(http.StatusConflict)
		writeJSONBody(&respWriter, resp.Message{"User with this email already exists"})
		return
	}

	UserService.UpdateUser(&userInfo)

	respWriter.WriteHeader(http.StatusOK)
	writeJSONBody(&respWriter, userInfo)

}

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

	// OMG really?
	// forum.Posts = PostService.CountOnForum(forum)

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

func ThreadDetails(respWriter http.ResponseWriter, request *http.Request) {
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

// does not perf
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

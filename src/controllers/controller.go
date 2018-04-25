package controllers

import (
	"encoding/json"
	"fmt"
	_ "fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/joeshaw/iso8601"

	"github.com/Alex-Kuz/tp-database/src/models"
	"github.com/Alex-Kuz/tp-database/src/router"
	"github.com/Alex-Kuz/tp-database/src/services"
	resp "github.com/Alex-Kuz/tp-database/src/utils/responses"
)


var (
	UserService   services.UserService
	ForumService  services.ForumService
	ThreadService services.ThreadService
)

const (
	POST = "POST"
	GET  = "GET"
)

func writeJsonBody(respWriter *http.ResponseWriter, v interface{}) {
	if err := json.NewEncoder(*respWriter).Encode(v); err != nil {
		(*respWriter).WriteHeader(500)
	}
}


func CreateUser(respWriter http.ResponseWriter, request *http.Request) {
	respWriter.Header().Set("Content-Type", "application/json; charset=utf-8")

	nickname := mux.Vars(request)["nickname"]

	user := models.User{}

	if err := json.NewDecoder(request.Body).Decode(&user); err != nil {
		panic(err)
	}
	user.Nickname = nickname

	scs, usersArray := UserService.AddUser(&user)


	if scs {
		respWriter.WriteHeader(http.StatusCreated)
		writeJsonBody(&respWriter, user)
	} else {
		respWriter.WriteHeader(http.StatusConflict)
		writeJsonBody(&respWriter, usersArray)
	}
}

func CreateForum(respWriter http.ResponseWriter, request *http.Request) {
	respWriter.Header().Set("Content-Type", "application/json; charset=utf-8")


	forum := models.Forum{}

	if err := json.NewDecoder(request.Body).Decode(&forum); err != nil {
		panic(err)
	}

	authorNickname := UserService.GetUserIDByNickname(forum.User)
	if authorNickname == nil {
		fmt.Println("CreateForum:  authorId = nil")

		respWriter.WriteHeader(http.StatusNotFound)
		writeJsonBody(&respWriter, resp.Message{"Forum master not found"})
		return
	}
	forum.User = *authorNickname

	scs, conflictForum := ForumService.AddForum(&forum)

	if scs {
		respWriter.WriteHeader(http.StatusCreated)
		writeJsonBody(&respWriter, *conflictForum)
	} else {
		respWriter.WriteHeader(http.StatusConflict)
		writeJsonBody(&respWriter, *conflictForum)
	}
}

func UserProfile(respWriter http.ResponseWriter, request *http.Request) {
	respWriter.Header().Set("Content-Type", "application/json; charset=utf-8")

	nickname := mux.Vars(request)["nickname"]

	user := UserService.GetUserByNickname(nickname)

	if user != nil {
		respWriter.WriteHeader(http.StatusOK)
		writeJsonBody(&respWriter, *user)
	} else {
		respWriter.WriteHeader(http.StatusNotFound)
		writeJsonBody(&respWriter, resp.MsgCantFindUser(nickname))
	}
}

func UpdateUser(respWriter http.ResponseWriter, request *http.Request) {
	respWriter.Header().Set("Content-Type", "application/json; charset=utf-8")

	nickname := mux.Vars(request)["nickname"]

	user := UserService.GetUserByNickname(nickname)
	if user == nil {
		respWriter.WriteHeader(http.StatusNotFound)
		writeJsonBody(&respWriter, resp.MsgCantFindUser(nickname))
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
	if opponent := UserService.GetUserByEmail(userInfo.Email);
			opponent != nil && *opponent != *user {
		respWriter.WriteHeader(http.StatusConflict)
		writeJsonBody(&respWriter, resp.Message{"User with this email already exists"})
		return
	}

	UserService.UpdateUser(&userInfo)

	respWriter.WriteHeader(http.StatusOK)
	writeJsonBody(&respWriter, userInfo)

}

func ForumDetails(respWriter http.ResponseWriter, request *http.Request) {
	respWriter.Header().Set("Content-Type", "application/json; charset=utf-8")

	slug := mux.Vars(request)["slug"]

	fmt.Println("ForumDetails: slug =", slug)

	forum := ForumService.GetForumBySlug(slug)
	if forum == nil {
		respWriter.WriteHeader(http.StatusNotFound)
		writeJsonBody(&respWriter, resp.Message{"Forum not found"})
		return
	}

	respWriter.WriteHeader(http.StatusOK)
	writeJsonBody(&respWriter, *forum)
}

func CreateThread(respWriter http.ResponseWriter, request *http.Request) {
	respWriter.Header().Set("Content-Type", "application/json; charset=utf-8")

	slug := mux.Vars(request)["slug"]

	thread := models.Thread{}
	if err := json.NewDecoder(request.Body).Decode(&thread); err != nil {
		panic(err)
	}
	thread.Forum = slug

	if author := UserService.GetUserByNickname(thread.Author); author == nil {
		respWriter.WriteHeader(http.StatusNotFound)
		writeJsonBody(&respWriter, resp.Message{"Thread author not found"})
		return
	}

	forumSlug := ForumService.SlugBySlug(thread.Forum)
	if forumSlug == nil {
		respWriter.WriteHeader(http.StatusNotFound)
		writeJsonBody(&respWriter, resp.Message{"Thread not found"})
		return
	}

	thread.Forum = *forumSlug

	if thread.Slug != "" {
		fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
		if anotherThread := ThreadService.GetThreadBySlug(thread.Slug); anotherThread != nil {
			respWriter.WriteHeader(http.StatusConflict)
			writeJsonBody(&respWriter, resp.Message{"Thread with same slug already exists"})
			return
		}
	}

	if len(thread.Created) == 0 {
		thread.Created = time.Now().UTC().Format(time.RFC3339)
		fmt.Println("\nCreateThread: thread.Created =", thread.Created)
	}


	fmt.Println("CreateThread: thread:", thread)

	ThreadService.AddThread(&thread)
	ForumService.IncThreadsCountBySlug(slug)

	respWriter.WriteHeader(http.StatusCreated)
	writeJsonBody(&respWriter, thread)
}

func ForumThreads(respWriter http.ResponseWriter, request *http.Request) {
	respWriter.Header().Set("Content-Type", "application/json; charset=utf-8")

	slug := mux.Vars(request)["slug"]

	forum := ForumService.GetForumBySlug(slug)
	if forum == nil {
		fmt.Println("CreateForum: forum with slug '", slug, "' not found")

		respWriter.WriteHeader(http.StatusNotFound)
		writeJsonBody(&respWriter, resp.Message{"Forum master not found"})
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

	_, threads := ThreadService.SelectThreads(slug, limit, since, desc)

	respWriter.WriteHeader(http.StatusOK)
	writeJsonBody(&respWriter, threads)
}


func MakeForumAPI(pgdb *services.PostgresDatabase) router.ForumAPI {
	forumAPI := make(router.ForumAPI)

	UserService = services.MakeUserService(pgdb)
	ForumService = services.MakeForumService(pgdb)
	ThreadService = services.MakeThreadService(pgdb)

	forumAPI["CreateUser"] = router.Route {
		Name:        "CreateUser",
		Method:      POST,
		Pattern:     "/user/{nickname}/create",
		HandlerFunc: CreateUser,
	}

	forumAPI["UserProfile"] = router.Route {
		Name:        "UserProfile",
		Method:      GET,
		Pattern:     "/user/{nickname}/profile",
		HandlerFunc: UserProfile,
	}

	forumAPI["UpdateUser"] = router.Route {
		Name:        "UpdateUser",
		Method:      POST,
		Pattern:     "/user/{nickname}/profile",
		HandlerFunc: UpdateUser,
	}

	forumAPI["CreateForum"] = router.Route {
		Name:        "CreateForum",
		Method:      POST,
		Pattern:     "/forum/create",
		HandlerFunc: CreateForum,
	}

	forumAPI["ForumDetails"] = router.Route {
		Name:        "ForumDetails",
		Method:      GET,
		Pattern:     "/forum/{slug}/details",
		HandlerFunc: ForumDetails,
	}

	forumAPI["CreateThread"] = router.Route {
		Name:        "CreateThread",
		Method:      POST,
		Pattern:     "/forum/{slug}/create",
		HandlerFunc: CreateThread,
	}

	forumAPI["ForumThreads"] = router.Route {
		Name:        "ForumThreads",
		Method:      GET,
		Pattern:     "/forum/{slug}/threads",
		HandlerFunc: ForumThreads,
	}

	return forumAPI
}



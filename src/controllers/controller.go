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
	PostService   services.PostService
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

	fmt.Println("\n----------------------------------------------------------------------------")


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
		if anotherThread := ThreadService.GetThreadBySlug(thread.Slug); anotherThread != nil {
			respWriter.WriteHeader(http.StatusConflict)
			writeJsonBody(&respWriter, *anotherThread)
			return
		}
	}

	if len(thread.Created) == 0 {
		thread.Created = time.Now().UTC().Format(time.RFC3339)
		fmt.Println("\nCreateThread: thread.Created =", thread.Created)
	}


	fmt.Println("CreateThread: thread{slug, created, author}:",
		thread.Slug, thread.Created, thread.Author)

	ThreadService.AddThread(&thread)
	ForumService.IncThreadsCountBySlug(slug)

	respWriter.WriteHeader(http.StatusCreated)
	writeJsonBody(&respWriter, thread)

	fmt.Println("----------------------------------------------------------------------------\n")
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
		writeJsonBody(&respWriter, resp.Message{"Forum not found"})
		return
	}

	respWriter.WriteHeader(http.StatusOK)
	writeJsonBody(&respWriter, *thread)
}

func CreatePosts(respWriter http.ResponseWriter, request *http.Request) {
	respWriter.Header().Set("Content-Type", "application/json; charset=utf-8")

	threadSlug := mux.Vars(request)["slug_or_id"]

	var thread *models.Thread
	threadId, err := strconv.ParseUint(threadSlug, 10, 64)
	if err == nil {
		thread = ThreadService.GetThreadById(threadId)
	} else {
		thread = ThreadService.GetThreadBySlug(threadSlug)
	}

	fmt.Println("\n----------------------------------------------------------------------------")

	fmt.Println("CreatePost: slug_or_id", threadSlug, err)

	if thread == nil {
		fmt.Println("CreatePost: thread with slug_or_id '", threadSlug, "' not found")

		respWriter.WriteHeader(http.StatusNotFound)
		writeJsonBody(&respWriter, resp.Message{"Thread not found"})
		fmt.Println("----------------------------------------------------------------------------\n")

		return
	}

	postsArray := make(models.PostsArray, 0)

	if err := json.NewDecoder(request.Body).Decode(&postsArray); err != nil {
		panic(err)
	}

	fmt.Println("CreatePost: posts:")
	for i := 0; i < len(postsArray); i++ {
		fmt.Println("\t", i, ":", postsArray[i])
	}

	parents := PostService.RequiredParents(postsArray)

	for i := 0; i < len(parents); i++ {
		if parent := PostService.GetPostById(parents[i]); parent == nil {
			respWriter.WriteHeader(http.StatusConflict)
			writeJsonBody(&respWriter, resp.Message{"Parents are not found"})
			return
		}
	}

	timeMoment := time.Now().UTC().Format(time.RFC3339)
	threadId = thread.ID
	forumSlug := thread.Forum
	for i := 0; i < len(postsArray); i++ {
		postsArray[i].Created = timeMoment
		postsArray[i].Thread = threadId
		postsArray[i].Forum = forumSlug

		fmt.Println("\t", i, ":", postsArray[i])

		if user := UserService.GetUserByNickname(postsArray[i].Author); user == nil {
			respWriter.WriteHeader(http.StatusNotFound)
			writeJsonBody(&respWriter, resp.Message{"Author are not found"})
			return
		}

		PostService.AddPost(&postsArray[i])
	}

	////////////////////////////

	respWriter.WriteHeader(http.StatusCreated)
	writeJsonBody(&respWriter, postsArray)

	fmt.Println("----------------------------------------------------------------------------\n")
}

func ThreadVote(respWriter http.ResponseWriter, request *http.Request) {
	respWriter.Header().Set("Content-Type", "application/json; charset=utf-8")

	threadSlug := mux.Vars(request)["slug_or_id"]

	var thread *models.Thread
	threadId, err := strconv.ParseUint(threadSlug, 10, 64)
	if err == nil {
		thread = ThreadService.GetThreadById(threadId)
	} else {
		thread = ThreadService.GetThreadBySlug(threadSlug)
	}

	fmt.Println("\n----------------------------------------------------------------------------")

	fmt.Println("ThreadVote: slug_or_id", threadSlug, err)

	if thread == nil {
		fmt.Println("ThreadVote: thread with slug_or_id '", threadSlug, "' not found")

		respWriter.WriteHeader(http.StatusNotFound)
		writeJsonBody(&respWriter, resp.Message{"Thread not found"})
		fmt.Println("----------------------------------------------------------------------------\n")

		return
	}

	var vote models.Vote
	if err := json.NewDecoder(request.Body).Decode(&vote); err != nil {
		panic(err)
	}


	thread = ThreadService.Vote(thread, vote)

	respWriter.WriteHeader(http.StatusOK)
	writeJsonBody(&respWriter, thread)
}

func MakeForumAPI(pgdb *services.PostgresDatabase) router.ForumAPI {
	forumAPI := make(router.ForumAPI)

	UserService = services.MakeUserService(pgdb)
	ForumService = services.MakeForumService(pgdb)
	ThreadService = services.MakeThreadService(pgdb)
	PostService = services.MakePostService(pgdb)

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

	forumAPI["ThreadDetails"] = router.Route {
		Name:        "ThreadDetails",
		Method:      GET,
		Pattern:     "/thread/{slug_or_id}/details",
		HandlerFunc: ThreadDetails,
	}

	forumAPI["CreatePosts"] = router.Route {
		Name:        "CreatePosts",
		Method:      POST,
		Pattern:     "/thread/{slug_or_id}/create",
		HandlerFunc: CreatePosts,
	}

	forumAPI["ThreadVote"] = router.Route {
		Name:        "ThreadVote",
		Method:      POST,
		Pattern:     "/thread/{slug_or_id}/vote",
		HandlerFunc: ThreadVote,
	}

	return forumAPI
}



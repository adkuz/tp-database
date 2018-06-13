package controllers

import (
	"encoding/json"
	_ "fmt"
	"net/http"
	"strconv"
	"strings"
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
		// fmt.Println("CreateForum:  authorId = nil")

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
	if opponent := UserService.GetUserByEmail(userInfo.Email); opponent != nil && *opponent != *user {
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

	forum := ForumService.GetForumBySlug(slug)
	if forum == nil {
		respWriter.WriteHeader(http.StatusNotFound)
		writeJsonBody(&respWriter, resp.Message{"Forum not found"})
		return
	}

	forum.Posts = PostService.CountOnForum(forum)

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
		if anotherThread := ThreadService.GetThreadBySlug(thread.Slug); anotherThread != nil {
			respWriter.WriteHeader(http.StatusConflict)
			writeJsonBody(&respWriter, *anotherThread)
			return
		}
	}

	if len(thread.Created) == 0 {
		thread.Created = time.Now().UTC().Format(time.RFC3339)
	}

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

	if thread == nil {
		respWriter.WriteHeader(http.StatusNotFound)
		writeJsonBody(&respWriter, resp.Message{"Thread not found"})
		return
	}

	postsArray := make(models.PostsArray, 0)
	if err := json.NewDecoder(request.Body).Decode(&postsArray); err != nil {
		panic(err)
	}

	parentsToThreads := PostService.RequiredParents(postsArray)

	timeMoment := time.Now().UTC().Format(time.RFC3339)
	threadId = thread.ID
	forumSlug := thread.Forum

	for i := 0; i < len(postsArray); i++ {
		postsArray[i].Created = timeMoment
		postsArray[i].Thread = threadId
		postsArray[i].Forum = forumSlug

		// TODO: many to one
		if _, ok := parentsToThreads[postsArray[i].Parent]; ok {
			if parent := PostService.GetPostById(postsArray[i].Parent); parent == nil {
				respWriter.WriteHeader(http.StatusConflict)
				writeJsonBody(&respWriter, resp.Message{"Parents are not found"})
				return
			} else if parent.Thread != threadId {
				respWriter.WriteHeader(http.StatusConflict)
				writeJsonBody(&respWriter, resp.Message{"Parent post was created in another thread"})
				return
			}
		}

		if user := UserService.GetUserByNickname(postsArray[i].Author); user == nil {
			respWriter.WriteHeader(http.StatusNotFound)
			writeJsonBody(&respWriter, resp.Message{"Author are not found"})
			return
		}

		PostService.AddPost(&postsArray[i])
	}

	ForumService.IncrementPostsCountBySlug(thread.Forum, len(postsArray))

	respWriter.WriteHeader(http.StatusCreated)
	writeJsonBody(&respWriter, postsArray)
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

	if thread == nil {
		respWriter.WriteHeader(http.StatusNotFound)
		writeJsonBody(&respWriter, resp.Message{"Thread not found"})
		return
	}

	var vote models.Vote
	if err := json.NewDecoder(request.Body).Decode(&vote); err != nil {
		panic(err)
	}

	user := UserService.GetUserByNickname(vote.Nickname)
	if user == nil {
		respWriter.WriteHeader(http.StatusNotFound)
		writeJsonBody(&respWriter, resp.MsgCantFindUser(vote.Nickname))
		return
	}

	thread = ThreadService.Vote(thread, vote)
	respWriter.WriteHeader(http.StatusOK)
	writeJsonBody(&respWriter, thread)
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
		writeJsonBody(&respWriter, resp.Message{"Thread not found"})
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
	writeJsonBody(&respWriter, posts)
}

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
		writeJsonBody(&respWriter, resp.Message{"Forum not found"})
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
	writeJsonBody(&respWriter, *thread)
}

func ForumUsers(respWriter http.ResponseWriter, request *http.Request) {
	respWriter.Header().Set("Content-Type", "application/json; charset=utf-8")

	threadSlug := mux.Vars(request)["slug"]

	forum := ForumService.GetForumBySlug(threadSlug)

	if forum == nil {
		respWriter.WriteHeader(http.StatusNotFound)
		writeJsonBody(&respWriter, resp.Message{"forum not found"})
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
	writeJsonBody(&respWriter, users)
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
		writeJsonBody(&respWriter, resp.Message{"Post not found"})
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
	writeJsonBody(&respWriter, postInfo)
}

func PostUpdate(respWriter http.ResponseWriter, request *http.Request) {
	respWriter.Header().Set("Content-Type", "application/json; charset=utf-8")

	id, err := strconv.ParseUint(mux.Vars(request)["id"], 10, 64)
	if err != nil {
		panic(err)
	}

	post := PostService.GetPostById(id)

	if post == nil {
		respWriter.WriteHeader(http.StatusNotFound)
		writeJsonBody(&respWriter, resp.Message{"Post not found"})
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
	writeJsonBody(&respWriter, post)
}

func ServiceStatus(respWriter http.ResponseWriter, request *http.Request) {
	respWriter.Header().Set("Content-Type", "application/json; charset=utf-8")

	status := make(map[string]uint64)

	count := func(tablename string) uint64 {

		rows := UserService.GetDB().Query("SELECT COUNT(*) FROM " + tablename)
		defer rows.Close()

		for rows.Next() {
			var count uint64
			err := rows.Scan(&count)
			if err != nil {
				panic(err)
			}
			return count
		}
		return 0
	}

	status["user"] = count(UserService.TableName())
	status["forum"] = count(ForumService.TableName())
	status["thread"] = count(ThreadService.TableName())
	status["post"] = count(PostService.TableName())

	respWriter.WriteHeader(http.StatusOK)
	writeJsonBody(&respWriter, status)
}

func ServiceClear(respWriter http.ResponseWriter, request *http.Request) {
	respWriter.Header().Set("Content-Type", "application/json; charset=utf-8")

	drop := func(tablename string) uint64 {

		rows := UserService.GetDB().Query("TRUNCATE TABLE " + tablename)
		defer rows.Close()

		for rows.Next() {
			var count uint64
			err := rows.Scan(&count)
			if err != nil {
				panic(err)
			}
			return count
		}
		return 0
	}

	drop("users, forums, threads, posts, votes, forum_users")

	respWriter.WriteHeader(http.StatusOK)
}

func MakeForumAPI(pgdb *services.PostgresDatabase) router.RouterAPI {
	forumAPI := make(router.RouterAPI)

	UserService = services.MakeUserService(pgdb)
	ForumService = services.MakeForumService(pgdb)
	ThreadService = services.MakeThreadService(pgdb)
	PostService = services.MakePostService(pgdb)

	forumAPI["CreateUser"] = router.Route{
		Name:        "CreateUser",
		Method:      POST,
		Pattern:     "/user/{nickname}/create",
		HandlerFunc: CreateUser,
	}

	forumAPI["UserProfile"] = router.Route{
		Name:        "UserProfile",
		Method:      GET,
		Pattern:     "/user/{nickname}/profile",
		HandlerFunc: UserProfile,
	}

	forumAPI["UpdateUser"] = router.Route{
		Name:        "UpdateUser",
		Method:      POST,
		Pattern:     "/user/{nickname}/profile",
		HandlerFunc: UpdateUser,
	}

	forumAPI["CreateForum"] = router.Route{
		Name:        "CreateForum",
		Method:      POST,
		Pattern:     "/forum/create",
		HandlerFunc: CreateForum,
	}

	forumAPI["ForumDetails"] = router.Route{
		Name:        "ForumDetails",
		Method:      GET,
		Pattern:     "/forum/{slug}/details",
		HandlerFunc: ForumDetails,
	}

	forumAPI["CreateThread"] = router.Route{
		Name:        "CreateThread",
		Method:      POST,
		Pattern:     "/forum/{slug}/create",
		HandlerFunc: CreateThread,
	}

	forumAPI["ForumThreads"] = router.Route{
		Name:        "ForumThreads",
		Method:      GET,
		Pattern:     "/forum/{slug}/threads",
		HandlerFunc: ForumThreads,
	}

	forumAPI["ThreadDetailsGet"] = router.Route{
		Name:        "ThreadDetails",
		Method:      GET,
		Pattern:     "/thread/{slug_or_id}/details",
		HandlerFunc: ThreadDetails,
	}

	forumAPI["ThreadUpdate"] = router.Route{
		Name:        "ThreadUpdate",
		Method:      POST,
		Pattern:     "/thread/{slug_or_id}/details",
		HandlerFunc: ThreadUpdate,
	}

	forumAPI["CreatePosts"] = router.Route{
		Name:        "CreatePosts",
		Method:      POST,
		Pattern:     "/thread/{slug_or_id}/create",
		HandlerFunc: CreatePosts,
	}

	forumAPI["ThreadVote"] = router.Route{
		Name:        "ThreadVote",
		Method:      POST,
		Pattern:     "/thread/{slug_or_id}/vote",
		HandlerFunc: ThreadVote,
	}

	forumAPI["ThreadPosts"] = router.Route{
		Name:        "ThreadPosts",
		Method:      GET,
		Pattern:     "/thread/{slug_or_id}/posts",
		HandlerFunc: ThreadPosts,
	}

	forumAPI["ForumUsers"] = router.Route{
		Name:        "ForumUsers",
		Method:      GET,
		Pattern:     "/forum/{slug}/users",
		HandlerFunc: ForumUsers,
	}

	forumAPI["PostDetails"] = router.Route{
		Name:        "PostDetails",
		Method:      GET,
		Pattern:     "/post/{id}/details",
		HandlerFunc: PostDetails,
	}

	forumAPI["PostUpdate"] = router.Route{
		Name:        "PostUpdate",
		Method:      POST,
		Pattern:     "/post/{id}/details",
		HandlerFunc: PostUpdate,
	}

	forumAPI["ServiceStatus"] = router.Route{
		Name:        "ServiceStatus",
		Method:      GET,
		Pattern:     "/service/status",
		HandlerFunc: ServiceStatus,
	}

	forumAPI["ServiceClear"] = router.Route{
		Name:        "ServiceClear",
		Method:      POST,
		Pattern:     "/service/clear",
		HandlerFunc: ServiceClear,
	}

	return forumAPI
}

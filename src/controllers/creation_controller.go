package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Alex-Kuz/tp-database/src/models"
	resp "github.com/Alex-Kuz/tp-database/src/utils/responses"
	"github.com/gorilla/mux"
)

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
		writeJSONBody(&respWriter, user)
	} else {
		respWriter.WriteHeader(http.StatusConflict)
		writeJSONBody(&respWriter, usersArray)
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
		writeJSONBody(&respWriter, resp.Message{"Forum master not found"})
		return
	}
	forum.User = *authorNickname

	scs, conflictForum := ForumService.AddForum(&forum)

	if scs {
		respWriter.WriteHeader(http.StatusCreated)
		writeJSONBody(&respWriter, *conflictForum)
	} else {
		respWriter.WriteHeader(http.StatusConflict)
		writeJSONBody(&respWriter, *conflictForum)
	}
}

// need to opt
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
		writeJSONBody(&respWriter, resp.Message{"Thread author not found"})
		return
	}

	forumSlug := ForumService.SlugBySlug(thread.Forum)
	if forumSlug == nil {
		respWriter.WriteHeader(http.StatusNotFound)
		writeJSONBody(&respWriter, resp.Message{"Thread not found"})
		return
	}

	thread.Forum = *forumSlug

	if thread.Slug != "" {
		if anotherThread := ThreadService.GetThreadBySlug(thread.Slug); anotherThread != nil {
			respWriter.WriteHeader(http.StatusConflict)
			writeJSONBody(&respWriter, *anotherThread)
			return
		}
	}

	if len(thread.Created) == 0 {
		thread.Created = time.Now().UTC().Format(time.RFC3339)
	}

	ThreadService.AddThread(&thread)
	ForumService.IncThreadsCountBySlug(slug)

	respWriter.WriteHeader(http.StatusCreated)
	writeJSONBody(&respWriter, thread)
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
		writeJSONBody(&respWriter, resp.Message{"Thread not found"})
		return
	}

	var vote models.Vote
	if err := json.NewDecoder(request.Body).Decode(&vote); err != nil {
		panic(err)
	}

	user := UserService.GetUserByNickname(vote.Nickname)
	if user == nil {
		respWriter.WriteHeader(http.StatusNotFound)
		writeJSONBody(&respWriter, resp.MsgCantFindUser(vote.Nickname))
		return
	}

	thread = ThreadService.Vote(thread, vote)

	respWriter.WriteHeader(http.StatusOK)
	writeJSONBody(&respWriter, thread)
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
		writeJSONBody(&respWriter, resp.Message{"Thread not found"})
		return
	}

	postsArray := make(models.PostsArray, 0)
	if err := json.NewDecoder(request.Body).Decode(&postsArray); err != nil {
		panic(err)
	}

	if len(postsArray) == 0 {
		respWriter.WriteHeader(http.StatusCreated)
		writeJSONBody(&respWriter, postsArray)
		return
	}

	timeMoment := time.Now().UTC().Format(time.RFC3339)
	threadId = thread.ID
	forumSlug := thread.Forum

	requiredAuthors := make(map[string]bool)
	for i := 0; i < len(postsArray); i++ {

		postsArray[i].Created = timeMoment
		postsArray[i].Thread = threadId
		postsArray[i].Forum = forumSlug

		if len(postsArray[i].Author) == 0 {
			respWriter.WriteHeader(http.StatusNotFound)
			writeJSONBody(&respWriter, resp.Message{"Null Author"})
			return
		}
		requiredAuthors[postsArray[i].Author] = true
	}

	realAuthors := UserService.GetUsersByNicknamesArray(requiredAuthors)
	if len(realAuthors) != len(requiredAuthors) {
		respWriter.WriteHeader(http.StatusNotFound)
		writeJSONBody(&respWriter, resp.Message{"Author are not found"})
		return
	}

	expectedParentsIDArray := PostService.RequiredParents(postsArray)
	success, postsArray := PostService.AddSomePosts(postsArray, expectedParentsIDArray)
	if !success {
		respWriter.WriteHeader(http.StatusConflict)
		writeJSONBody(&respWriter, resp.Message{"Parent post was created in another thread, or not found"})
		return
	}

	ForumService.IncrementPostsCountBySlug(thread.Forum, len(postsArray))
	ForumService.AddUsers(realAuthors, forumSlug)

	respWriter.WriteHeader(http.StatusCreated)
	writeJSONBody(&respWriter, postsArray)
}

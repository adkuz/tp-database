package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/Alex-Kuz/tp-database/src/router"
	"github.com/Alex-Kuz/tp-database/src/services"
)

// services
var (
	UserService   services.UserService
	ForumService  services.ForumService
	ThreadService services.ThreadService
	PostService   services.PostService
)

// methods
const (
	POST = "POST"
	GET  = "GET"
)

func writeJSONBody(respWriter *http.ResponseWriter, v interface{}) {
	if err := json.NewEncoder(*respWriter).Encode(v); err != nil {
		(*respWriter).WriteHeader(500)
	}
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

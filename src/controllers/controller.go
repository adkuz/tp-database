package controllers

import (
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"

	"github.com/Alex-Kuz/tp-database/src/services"
	"github.com/Alex-Kuz/tp-database/src/models"
	"github.com/Alex-Kuz/tp-database/src/router"
	resp "github.com/Alex-Kuz/tp-database/src/utils/responses"
)


var (
	UserService services.UserService
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

// url := "/lolkek/{nickname}/shit"
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

	userInfo := models.User{}
	if err := json.NewDecoder(request.Body).Decode(&userInfo); err != nil {
		panic(err)
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


func MakeForumAPI(pgdb *services.PostgresDatabase) router.ForumAPI {
	forumAPI := make(router.ForumAPI)

	UserService = services.MakeUserService(pgdb)

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

	return forumAPI
}



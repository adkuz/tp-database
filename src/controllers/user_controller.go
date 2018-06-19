package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/Alex-Kuz/tp-database/src/models"
	resp "github.com/Alex-Kuz/tp-database/src/utils/responses"
	"github.com/gorilla/mux"
)

// optimized with index
func UserProfile(respWriter http.ResponseWriter, request *http.Request) {
	respWriter.Header().Set("Content-Type", "application/json; charset=utf-8")

	nickname := mux.Vars(request)["nickname"]

	user := UserService.GetUserByNickname(nickname)

	if user == nil {
		respWriter.WriteHeader(http.StatusNotFound)
		writeJSONBody(&respWriter, resp.MsgCantFindUser(nickname))
		return
	}

	respWriter.WriteHeader(http.StatusOK)
	writeJSONBody(&respWriter, *user)
}

// does not perf
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

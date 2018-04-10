package responses

import (
	"fmt"
)

const (
	CantFindUser = "Can't find user with nickname '%s'"
)

type Message struct {
	Msg string `json:"message"`
}

func MsgCantFindUser(nickname string) Message {
	return Message{
		Msg: fmt.Sprintf(CantFindUser, nickname),
	}
}

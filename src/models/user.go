package models

import (
	"github.com/go-openapi/swag"
	"github.com/goware/emailx"
)

type ValidError int

const (
	ALL_RIGHT ValidError = iota
	INVALID_EMAIL
	INVALID_NICKNAME
	NIL_USER
)


type User struct {

	About    string `json:"about" db:"about"`

	Email    string `json:"email" db:"email"`

	Fullname string `json:"fullname" db:"fullname"`

	Nickname string `json:"nickname" db:"nickname"`
}

type UsersArray []User


func IsValid(user *User) ValidError {
	if user == nil {
		return NIL_USER
	}

	if err := emailx.Validate((*user).Email); err != nil {
		return INVALID_EMAIL
	}

	return ALL_RIGHT
}


// MarshalBinary interface implementation
func (m *User) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *User) UnmarshalBinary(b []byte) error {
	var res User
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}

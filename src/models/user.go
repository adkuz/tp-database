package models

import (
	"github.com/go-openapi/swag"
)

type User struct {
	About string `json:"about"`

	Email string `json:"email"`

	Fullname string `json:"fullname"`

	Nickname string `json:"nickname"`
}

type UsersArray []User

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

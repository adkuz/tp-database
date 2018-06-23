package models

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/swag"
)

type PostDateTime *strfmt.DateTime

type Post struct {
	Author string `json:"author"`

	Created string `json:"created,omitempty"`

	Forum string `json:"forum,omitempty"`

	ID uint64 `json:"id,omitempty"`

	IsEdited bool `json:"isEdited,omitempty"`

	Message string `json:"message"`

	Parent uint64 `json:"parent,omitempty"`

	Thread uint64 `json:"thread,omitempty"`

	Path []uint64 `json:",omitempty"`
}

type PostsArray []Post

// MarshalBinary interface implementation
func (m *Post) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Post) UnmarshalBinary(b []byte) error {
	var res Post
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}

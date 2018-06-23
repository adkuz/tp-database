package models

import (
	"github.com/go-openapi/swag"
)

//type ThreadDateTime *strfmt.DateTime

type Thread struct {
	Author string `json:"author"`

	Created string `json:"created,omitempty"`

	Forum string `json:"forum,omitempty"`

	ID uint64 `json:"id,omitempty"`

	Message string `json:"message"`

	Slug string `json:"slug,omitempty"`

	Title string `json:"title"`

	Votes int64 `json:"votes,omitempty"`
}

type ThreadsArray []Thread

// MarshalBinary interface implementation
func (m *Thread) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Thread) UnmarshalBinary(b []byte) error {
	var res Thread
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}

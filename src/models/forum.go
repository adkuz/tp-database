package models

import (
	"github.com/go-openapi/swag"
)

type Forum struct {
	Posts uint64 `json:"posts,omitempty"`

	Slug string `json:"slug"`

	Threads uint32 `json:"threads,omitempty"`

	Title string `json:"title"`

	User string `json:"user"`
}

type ForumsArray []Forum

// MarshalBinary interface implementation
func (m *Forum) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Forum) UnmarshalBinary(b []byte) error {
	var res Forum
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}

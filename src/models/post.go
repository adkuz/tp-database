package models


import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

type DateTime *strfmt.DateTime


type Post struct {

	Author  string   `json:"author"`

	Created DateTime `json:"created,omitempty"`

	Forum   string   `json:"forum,omitempty"`

	ID      int64    `json:"id,omitempty"`

	IsEdited bool    `json:"isEdited,omitempty"`

	Message  string  `json:"message"`

	Parent   int64   `json:"parent,omitempty"`

	Thread   int32   `json:"thread,omitempty"`
}


type PostsArray []Post



func (m *Post) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateAuthor(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateMessage(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Post) validateAuthor(formats strfmt.Registry) error {

	if err := validate.RequiredString("author", "body", string(m.Author)); err != nil {
		return err
	}

	return nil
}

func (m *Post) validateMessage(formats strfmt.Registry) error {

	if err := validate.RequiredString("message", "body", string(m.Message)); err != nil {
		return err
	}

	return nil
}

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

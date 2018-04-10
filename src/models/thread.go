package models


import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)


type Thread struct {

	Author  string `json:"author"`

	Created *strfmt.DateTime `json:"created,omitempty"`

	Forum   string `json:"forum,omitempty"`

	ID      int32  `json:"id,omitempty"`

	Message string `json:"message"`

	// Pattern: ^(\d|\w|-|_)*(\w|-|_)(\d|\w|-|_)*$
	Slug    string `json:"slug,omitempty"`

	Title   string `json:"title"`

	Votes   int32  `json:"votes,omitempty"`
}


type ThreadsArray []Thread


func (m *Thread) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateAuthor(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateMessage(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateSlug(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateTitle(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Thread) validateAuthor(formats strfmt.Registry) error {

	if err := validate.RequiredString("author", "body", string(m.Author)); err != nil {
		return err
	}

	return nil
}

func (m *Thread) validateMessage(formats strfmt.Registry) error {

	if err := validate.RequiredString("message", "body", string(m.Message)); err != nil {
		return err
	}

	return nil
}

func (m *Thread) validateSlug(formats strfmt.Registry) error {

	if swag.IsZero(m.Slug) { // not required
		return nil
	}

	if err := validate.Pattern("slug", "body", string(m.Slug), `^(\d|\w|-|_)*(\w|-|_)(\d|\w|-|_)*$`); err != nil {
		return err
	}

	return nil
}

func (m *Thread) validateTitle(formats strfmt.Registry) error {

	if err := validate.RequiredString("title", "body", string(m.Title)); err != nil {
		return err
	}

	return nil
}

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

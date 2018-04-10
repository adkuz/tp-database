package models


import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)


type Forum struct {

	Posts   int64  `json:"posts,omitempty"`

	// Pattern: ^(\d|\w|-|_)*(\w|-|_)(\d|\w|-|_)*$
	Slug    string `json:"slug"`

	Threads int32  `json:"threads,omitempty"`

	Title   string `json:"title"`

	User    string `json:"user"`
}


type ForumsArray []Forum


func (m *Forum) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateSlug(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateTitle(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateUser(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Forum) validateSlug(formats strfmt.Registry) error {

	if err := validate.RequiredString("slug", "body", string(m.Slug)); err != nil {
		return err
	}

	if err := validate.Pattern("slug", "body", string(m.Slug), `^(\d|\w|-|_)*(\w|-|_)(\d|\w|-|_)*$`); err != nil {
		return err
	}

	return nil
}

func (m *Forum) validateTitle(formats strfmt.Registry) error {

	if err := validate.RequiredString("title", "body", string(m.Title)); err != nil {
		return err
	}

	return nil
}

func (m *Forum) validateUser(formats strfmt.Registry) error {

	if err := validate.RequiredString("user", "body", string(m.User)); err != nil {
		return err
	}

	return nil
}

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

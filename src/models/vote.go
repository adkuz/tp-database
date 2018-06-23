package models

import (
	"encoding/json"

	"github.com/go-openapi/swag"
)

type Vote struct {
	Nickname string `json:"nickname"`

	Voice int32 `json:"voice"`
}

var voteTypeVoicePropEnum []interface{}

func init() {
	var res []int32
	if err := json.Unmarshal([]byte(`[-1,1]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		voteTypeVoicePropEnum = append(voteTypeVoicePropEnum, v)
	}
}

// MarshalBinary interface implementation
func (m *Vote) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Vote) UnmarshalBinary(b []byte) error {
	var res Vote
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}

package cep

import (
	"encoding/json"
	"strings"
)

// Address has address informations.
type Address struct {
	CEP          string `json:"cep"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	State        string `json:"state"`
	Street       string `json:"street"`
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (a *Address) UnmarshalJSON(data []byte) error {
	fields := make(map[string]string)
	err := json.Unmarshal(data, &fields)
	if err != nil {
		return &Error{Kind: UnmarshalErr, Err: err}
	}
	for name, val := range fields {
		switch strings.ToLower(name) {
		case "cep":
			a.CEP, err = Canonicalize(val)
			if err != nil {
				return &Error{Kind: UnmarshalErr, Err: err}
			}
		case "city", "localidade":
			a.City = val
		case "neighborhood", "bairro":
			a.Neighborhood = val
		case "state", "uf":
			a.State = val
		case "street", "logradouro":
			a.Street = val
		}
	}
	return nil
}

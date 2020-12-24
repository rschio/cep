package cep

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type viacep struct {
	client *http.Client
}

func (viacep) Name() string { return "viacep" }

func (v *viacep) Fetch(ctx context.Context, cep string) (Address, error) {
	url := fmt.Sprintf("https://viacep.com.br/ws/%s/json/unicode/", cep)
	payload, err := getReq(ctx, v.client, url)
	if err != nil {
		return Address{}, fmt.Errorf("%s: %w", v.Name(), err)
	}
	addr, err := unmarshalViacep(payload)
	if err != nil {
		return Address{}, fmt.Errorf("%s: %w", v.Name(), err)
	}
	return addr, nil
}

type viacepErr struct {
	Erro bool
}

func unmarshalViacep(payload []byte) (Address, error) {
	addr := new(Address)
	if err := json.Unmarshal(payload, addr); err != nil {
		// Failed to unmarshal the payload. It can mean
		// the payload is corrupted or something like
		// that or it can mean the CEP was not found
		// in the database.
		vcepErr := &viacepErr{}
		err1 := json.Unmarshal(payload, vcepErr)
		if err1 != nil || vcepErr.Erro != true {
			return Address{}, err
		}
		return Address{}, errors.New("CEP not found")
	}
	return *addr, nil
}

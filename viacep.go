package cep

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type viacep struct {
	client *http.Client
}

func (viacep) Name() string { return "viacep" }

// Viacep returns a fetcher of viacep.
func Viacep(c *http.Client) Fetcher {
	if c == nil {
		c = http.DefaultClient
	}
	return &viacep{client: c}
}

// Fetch implements the Fetcher interface.
func (v *viacep) Fetch(ctx context.Context, cep string) (Address, error) {
	url := fmt.Sprintf("https://viacep.com.br/ws/%s/json/unicode/", cep)
	payload, err := v.req(ctx, url)
	if err != nil {
		return Address{}, &Error{
			Kind: err.Kind,
			Err:  fmt.Errorf("%s: %s", v.Name(), err.Error()),
		}
	}
	addr, err := v.unmarshal(payload)
	if err != nil {
		return Address{}, &Error{
			Kind: err.Kind,
			Err:  fmt.Errorf("%s: %s", v.Name(), err.Error()),
		}
	}
	return addr, nil
}

func (v *viacep) req(ctx context.Context, url string) ([]byte, *Error) {
	rsp, err := request(ctx, v.client, url)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != 200 {
		return nil, &Error{Kind: Other, Err: errors.New(rsp.Status)}
	}
	data, err1 := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, &Error{Kind: Other, Err: err1}
	}
	return data, nil
}

type viacepErr struct {
	Erro bool
}

func (viacep) unmarshal(payload []byte) (Address, *Error) {
	addr := new(Address)
	if err := json.Unmarshal(payload, addr); err != nil {
		// Failed to unmarshal the payload. It can mean
		// the payload is corrupted or something like
		// that or it can mean the CEP was not found
		// in the database.
		vcepErr := &viacepErr{}
		err1 := json.Unmarshal(payload, vcepErr)
		if err1 != nil || vcepErr.Erro != true {
			return Address{}, &Error{
				Kind: UnmarshalErr,
				Err:  err,
			}
		}
		return Address{}, &Error{
			Kind: CEPNotFound,
			Err:  fmt.Errorf("CEP not found"),
		}
	}
	return *addr, nil
}

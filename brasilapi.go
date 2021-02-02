package cep

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type brasilapi struct {
	client *http.Client
}

func (brasilapi) Name() string { return "brasilapi" }

// Brasilapi returns a fetcher of brasilapi.
func Brasilapi(c *http.Client) Fetcher {
	if c == nil {
		c = http.DefaultClient
	}
	return &brasilapi{client: c}
}

// Fetch implements the Fetcher interface.
func (b *brasilapi) Fetch(ctx context.Context, cep string) (Address, error) {
	url := fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep)
	rsp, err := request(ctx, b.client, url)
	if err != nil {
		return Address{}, &Error{
			Kind: err.Kind,
			Err:  fmt.Errorf("%s: %s", b.Name(), err.Error()),
		}
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != 200 {
		// BrasilAPI returns 404 when the CEP was not found.
		if rsp.StatusCode != 404 {
			return Address{}, &Error{
				Kind: Other,
				Err:  fmt.Errorf("%s: %s", b.Name(), err.Error()),
			}
		}
		// TODO: Check if the 404 is in fact a CEP not found.
		// It can be a server problem...
		return Address{}, &Error{
			Kind: CEPNotFound,
			Err:  fmt.Errorf("%s: CEP not found", b.Name()),
		}
	}
	addr := new(Address)
	if err1 := json.NewDecoder(rsp.Body).Decode(addr); err1 != nil {
		return Address{}, &Error{
			Kind: UnmarshalErr,
			Err:  fmt.Errorf("%s: %s", b.Name(), err1.Error()),
		}
	}
	return *addr, nil
}

type timeouter interface {
	Timeout() bool
}

func request(ctx context.Context, c *http.Client, url string) (*http.Response, *Error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, &Error{
			Kind: Other,
			Err:  err,
		}
	}
	rsp, err := c.Do(req)
	if err != nil {
		e := &Error{Err: err}
		if terr, ok := err.(timeouter); ok && terr.Timeout() {
			e.Kind = Timeout
			return nil, e
		}
		e.Kind = Other
		return nil, e
	}
	return rsp, nil
}

package cep

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type brasilapi struct {
	client *http.Client
}

func (brasilapi) Name() string { return "brasilapi" }

func (b *brasilapi) Fetch(ctx context.Context, cep string) (Address, error) {
	url := fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep)
	rc, err := makeReq(ctx, b.client, url)
	if err != nil {
		return Address{}, fmt.Errorf("%s: %w", b.Name(), err)
	}
	defer rc.Close()

	addr := new(Address)
	if err = json.NewDecoder(rc).Decode(addr); err != nil {
		return Address{}, fmt.Errorf("%s: %w", b.Name(), err)
	}
	return *addr, nil
}

func makeReq(ctx context.Context, c *http.Client, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	rsp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if rsp.StatusCode != 200 {
		rsp.Body.Close()
		return nil, errors.New(rsp.Status)
	}
	return rsp.Body, nil
}

func getReq(ctx context.Context, c *http.Client, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	rsp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != 200 {
		return nil, errors.New(rsp.Status)
	}

	return ioutil.ReadAll(rsp.Body)
}

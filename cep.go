package cep

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
)

type Fetcher interface {
	Fetch(ctx context.Context, cep string) (Address, error)
}

type Client struct {
	fetchers []Fetcher
}

func NewClient(fs []Fetcher) *Client {
	if len(fs) == 0 {
		fs = []Fetcher{Brasilapi(nil), Viacep(nil)}
	}
	return &Client{fetchers: fs}
}

func Brasilapi(c *http.Client) Fetcher {
	if c == nil {
		c = http.DefaultClient
	}
	return &brasilapi{client: c}
}

func Viacep(c *http.Client) Fetcher {
	if c == nil {
		c = http.DefaultClient
	}
	return &viacep{client: c}
}

func (c *Client) Search(ctx context.Context, cep string) (Address, error) {
	if len(c.fetchers) == 0 {
		return Address{}, errors.New("client has no fetcher")
	}
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()

	var err error
	cep, err = Canonicalize(cep)
	if err != nil {
		return Address{}, err
	}

	addrc := make(chan Address)
	// The buffer here is important to not leak
	// the goroutines.
	errc := make(chan error, len(c.fetchers))

	var wg sync.WaitGroup
	wg.Add(len(c.fetchers))
	for _, f := range c.fetchers {
		go func(f Fetcher) {
			defer wg.Done()
			addr, err := f.Fetch(ctx, cep)
			if err != nil {
				errc <- err
				return
			}
			select {
			case addrc <- addr:
			case <-ctx.Done():
			}
		}(f)
	}
	// Signal all services failed
	// to find the CEP.
	go func() {
		wg.Wait()
		close(errc)
	}()
	// Store the first error to return if all services
	// failed to find the CEP. Use the first and not the
	// last one because the last one has more probability
	// to be a timeout error.
	var firstErr error
	for {
		select {
		case <-ctx.Done():
			return Address{}, ctx.Err()
		case addr := <-addrc:
			return addr, nil
		case err, ok := <-errc:
			if firstErr == nil {
				firstErr = err
			}
			if !ok {
				return Address{}, firstErr
			}
		}
	}
}

func Valid(cep string) error {
	_, err := Canonicalize(cep)
	return err
}

func Canonicalize(cep string) (string, error) {
	if cep == "" {
		return "", errors.New("invalid CEP")
	}
	cep = strings.TrimSpace(cep)
	// A CEP can contain a '-' that separates the
	// first 5 digits from the last 3. If the '-'
	// is in the rigth position, remove it.
	if p := strings.IndexByte(cep, '-'); p >= 0 {
		if len(cep)-p != 4 {
			return "", errors.New("wrong position of '-'")
		}
		cep = cep[:p] + cep[p+1:]
	}
	// Only the first digit can be ommited, so the
	// length needs to be 7 or 8.
	l := len(cep)
	if !(7 <= l && l <= 8) {
		return "", errors.New("invalid CEP length")
	}
	for _, r := range cep {
		if !('0' <= r && r <= '9') {
			return "", errors.New("illegal character")
		}
	}
	if l < 8 {
		cep = "0" + cep
	}
	return cep, nil
}

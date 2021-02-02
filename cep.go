package cep

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
)

// ErrKind is used to compare errors.
type ErrKind uint32

// Error is an error with more information.
type Error struct {
	Kind ErrKind
	Err  error
}

func (e *Error) append(errs ...error) {
	for _, err := range errs {
		if err == nil {
			continue
		}
		e1 := new(Error)
		if !errors.As(err, &e1) {
			e.Kind = Other
			e.Err = fmt.Errorf("%v\n\t%w", e.Err, err)
			continue
		}
		if e.Kind == 0 {
			e.Kind = e1.Kind
		}
		if e.Kind != e1.Kind {
			e.Kind = Other
		}
		if e.Err == nil {
			e.Err = e1.Err
		} else {
			e.Err = fmt.Errorf("%v\n\t%w", e.Err, e1)
		}
	}
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Kind == Other {
		return fmt.Sprintf("Other:\n\t%v", e.Err)
	}
	return fmt.Sprintf("%v", e.Kind)
}

// Is checks if target has the same Kind as e.
// Kind Other is not comparable so if e.Kind == Other
// or target.Kind == Other it returns false.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if t == nil || !ok {
		return false
	}
	// Other is not comparable.
	if e.Kind == Other || t.Kind == Other {
		return false
	}
	return e.Kind == t.Kind
}

const (
	CEPNotFound     ErrKind = 1 << iota // CEP Not Found
	ContextCanceled                     // Context Canceled
	InvalidCEP                          // Invalid CEP
	Timeout                             // Timeout
	UnmarshalErr                        // Unmarshal error

	// Other is not comparable.
	Other // Other
)

// Fetcher fetchs a cep.
type Fetcher interface {
	// Fetch searches for the Address of cep. Fetch must be
	// concurrent safe.
	Fetch(ctx context.Context, cep string) (Address, error)
}

// Client can search for an address using a CEP. A client can have
// many fetchers and use them concurrently to find the address.
type Client struct {
	fs []Fetcher
}

// NewClient returns a new client with the provided
// fetchers, if no fetcher was provided it uses the
// DefaultFetchers.
func NewClient(fetchers []Fetcher) *Client {
	return &Client{fs: fetchers}
}

// DefaultFetchers are the default fetchers used
// when no fetcher is provided.
var DefaultFetchers = []Fetcher{
	Brasilapi(nil),
	Viacep(nil),
}

func (c *Client) fetchers() []Fetcher {
	if len(c.fs) > 0 {
		return c.fs
	}
	return DefaultFetchers
}

// Search searches for the Address of CEP cep using client fetchers.
// It returns the first result without error or wait all the errors and return
// a zero Address and an error.
func (c *Client) Search(ctx context.Context, cep string) (Address, error) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()

	var err error
	cep, err = Canonicalize(cep)
	if err != nil {
		return Address{}, err
	}

	fetchers := c.fetchers()
	addrc := make(chan Address)
	// The buffer here is important to not leak
	// the goroutines.
	errc := make(chan error, len(fetchers))

	var wg sync.WaitGroup
	wg.Add(len(fetchers))
	for _, f := range fetchers {
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

	e := new(Error)
	for {
		select {
		case addr := <-addrc:
			return addr, nil
		case <-ctx.Done():
			kind := Other
			switch ctx.Err() {
			case context.Canceled:
				kind = ContextCanceled
			case context.DeadlineExceeded:
				kind = Timeout
			}
			e.append(&Error{Kind: kind, Err: ctx.Err()})
			return Address{}, e
		case err, ok := <-errc:
			e.append(err)
			if !ok {
				return Address{}, e
			}
		}
	}
}

// Valid verifies if a cep is valid.
// If it is valid, return nil.
func Valid(cep string) error {
	_, err := Canonicalize(cep)
	return err
}

// Canonicalize transforms the cep in it's canonical form
// that is 8 numbers without any slash. If the cep is not
// valid it returns an empty string and an error.
func Canonicalize(cep string) (string, error) {
	e := &Error{Kind: InvalidCEP}
	cep = strings.TrimSpace(cep)
	if cep == "" {
		e.Err = errors.New("empty")
		return "", e
	}
	// A CEP can contain a '-' that separates the
	// first 5 digits from the last 3. If the '-'
	// is in the rigth position, remove it.
	if p := strings.IndexByte(cep, '-'); p >= 0 {
		if len(cep)-p != 4 {
			e.Err = errors.New("wrong position of '-'")
			return "", e
		}
		cep = cep[:p] + cep[p+1:]
	}
	// Only the first digit can be ommitted, so the
	// length needs to be 7 or 8.
	l := len(cep)
	if !(7 <= l && l <= 8) {
		e.Err = errors.New("invalid CEP length")
		return "", e
	}
	for _, r := range cep {
		if !('0' <= r && r <= '9') {
			e.Err = errors.New("illegal character")
			return "", e
		}
	}
	if l < 8 {
		cep = "0" + cep
	}
	return cep, nil
}

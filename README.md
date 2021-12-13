# cep
[![Go Reference](https://pkg.go.dev/badge/github.com/rschio/cep.svg)](https://pkg.go.dev/github.com/rschio/cep)

Search for a brazilian address using the CEP. It can use more than one API to seach concurrently. 
This library is inspired by the cep-promise library.

### Features:
- [x] Comparable errors.
- [x] Possible to use custom fetchers.
- [x] Search concurrently.

### Examples:

#### Default Fetchers
```go
package main

import (
	"context"
	"fmt"

	"github.com/rschio/cep"
)

func main() {
	c := cep.NewClient(nil)
	addr, err := c.Search(context.TODO(), "01310000")
	if err != nil {
		// Handle err.
	}
	fmt.Printf("%#v", addr)
}

```

#### Custom Fetcher
```go
package main

import (
	"context"
	"fmt"

	"github.com/rschio/cep"
)

type myFetcher struct{}

func (myFetcher) Fetch(_ context.Context, CEP string) (cep.Address, error) {
	return cep.Address{
		CEP:          CEP,
		City:         "City",
		Neighborhood: "Neighborhood",
		State:        "State",
		Street:       "Street",
	}, nil
}

func main() {
	fetchers := []cep.Fetcher{myFetcher{}}
	c := cep.NewClient(fetchers)
	addr, err := c.Search(context.TODO(), "01310000")
	if err != nil {
		// Handle err.
	}
	fmt.Printf("%#v", addr)
}
```

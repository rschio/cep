package cep

import (
	"context"
	"net/http"
	"reflect"
	"testing"
)

func TestSearch(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		CEP      string
		Fetchers []Fetcher
		want     Address
	}{
		{"nil", "01310000", nil, addr},
		{"zero len", "01310000", []Fetcher{}, addr},
		{"brasilapi", "01310000", []Fetcher{Brasilapi(nil)}, addr},
		{"viacep", "01310000", []Fetcher{Viacep(&http.Client{})}, addr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			c := NewClient(tt.Fetchers)
			got, err := c.Search(ctx, tt.CEP)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(tt.want, got) {
				t.Fatalf("want: %#v\ngot:%#v\n", tt.want, got)
			}
		})
	}
}

func TestCanonicalize(t *testing.T) {
	tests := []struct {
		name string
		CEP  string
		want string
	}{
		{"numeric", "01310000", "01310000"},
		{"with '-'", "01310-000", "01310000"},
		{"without 0", "1310-000", "01310000"},
		{"wrong position of '-'", "01310000-", ""},
		{"wrong position 2 of '-'", "-01310000", ""},
		{"wrong position 3 of '-'", "013-10000", ""},
		{"wrong position 4 of '-'", "-1310000", ""},
		{"without 00", "310000", ""},
		{"letter", "12a34567", ""},
		{"letter 2", "a1234567", ""},
		{"UTF-8", "34㤹-678", ""},
		{"zero val", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Canonicalize(tt.CEP)
			if tt.want != got {
				t.Fatalf("want: %s, got: %s, err: %v", tt.want, got, err)
			}
		})
	}
}

func TestValid(t *testing.T) {
	tests := []struct {
		name string
		CEP  string
		want bool
	}{
		{"good", "1310000", true},
		{"bad", "01 310000", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Valid(tt.CEP)
			if tt.want != (err == nil) {
				t.Fatal("valid failed")
			}
		})
	}
}

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want Address
	}{
		{"viacep", viacepData, addr},
		{"brasilapi", brasilapiData, addr},
	}
	a := new(Address)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := a.UnmarshalJSON(tt.data)
			if err != nil || !reflect.DeepEqual(*a, tt.want) {
				t.Fatalf("got: %#v\nwant: %#v", *a, tt.want)
			}
		})
	}
}

var addr = Address{
	CEP:          "01310000",
	City:         "São Paulo",
	Neighborhood: "Bela Vista",
	State:        "SP",
	Street:       "Avenida Paulista",
}

var viacepData = []byte(`{
  "cep": "01310-000",
  "logradouro": "Avenida Paulista",
  "complemento": "até 610 - lado par",
  "bairro": "Bela Vista",
  "localidade": "São Paulo",
  "uf": "SP",
  "ibge": "3550308",
  "gia": "1004",
  "ddd": "11",
  "siafi": "7107"
}`)

var brasilapiData = []byte(`{
    "cep": "01310000",
    "state": "SP",
    "city": "São Paulo",
    "neighborhood": "Bela Vista",
    "street": "Avenida Paulista"
}`)

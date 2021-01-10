package cep

import (
	"context"
	"net/http"
	"testing"
)

func TestFetchViacep(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		CEP  string
		want bool
	}{
		{"good", "01310000", true},
		{"bad", "00000000", false},
	}
	ctx := context.Background()
	v := &viacep{client: http.DefaultClient}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := v.Fetch(ctx, tt.CEP)
			if tt.want != (err == nil) {
				t.Fatal("valid failed")
			}
		})
	}
}

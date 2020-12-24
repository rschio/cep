package cep

import (
	"context"
	"net/http"
	"testing"
)

func TestBrasilapiFetch(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		cep     string
		wantErr bool
	}{
		{"good", "01310000", false},
		{"bad", "00000000", true},
	}
	client := &brasilapi{client: http.DefaultClient}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, err := client.Fetch(ctx, tt.cep)
			if tt.wantErr != (err != nil) {
				t.Fatalf("wantErr: %v, got: %v", tt.wantErr, err)
			}
		})
	}
}

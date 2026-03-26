package grpc

import (
	"context"
	"testing"

	"google.golang.org/grpc/metadata"
)

func TestExtractTaskToken(t *testing.T) {
	t.Run("authorization header wins", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer header-token"))

		token, err := extractTaskToken(ctx)
		if err != nil {
			t.Fatalf("extractTaskToken() error = %v", err)
		}
		if token != "header-token" {
			t.Fatalf("extractTaskToken() = %q, want header-token", token)
		}
	})

	t.Run("missing token", func(t *testing.T) {
		if _, err := extractTaskToken(context.Background()); err == nil {
			t.Fatalf("extractTaskToken() error = nil, want non-nil")
		}
	})
}

func TestParseBearerToken(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
		ok    bool
	}{
		{name: "valid", value: "Bearer abc", want: "abc", ok: true},
		{name: "case insensitive", value: "bearer abc", want: "abc", ok: true},
		{name: "trim spaces", value: "  Bearer abc  ", want: "abc", ok: true},
		{name: "missing prefix", value: "abc", ok: false},
		{name: "empty token", value: "Bearer ", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseBearerToken(tt.value)
			if ok != tt.ok || got != tt.want {
				t.Fatalf("parseBearerToken(%q) = (%q, %v), want (%q, %v)", tt.value, got, ok, tt.want, tt.ok)
			}
		})
	}
}

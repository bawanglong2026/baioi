package safeurl

import (
	"context"
	"net"
	"testing"
)

type resolverStub map[string][]net.IPAddr

func (r resolverStub) LookupIPAddr(_ context.Context, host string) ([]net.IPAddr, error) {
	return r[host], nil
}

func TestValidatePublicHTTP(t *testing.T) {
	r := resolverStub{
		"public.example":  {{IP: net.ParseIP("93.184.216.34")}},
		"private.example": {{IP: net.ParseIP("10.0.0.5")}},
	}
	for _, tc := range []struct {
		raw string
		ok  bool
	}{
		{"https://public.example/image.png", true},
		{"http://127.0.0.1/a", false},
		{"http://169.254.169.254/latest/meta-data", false},
		{"http://private.example/a", false},
		{"file:///etc/passwd", false},
	} {
		if err := ValidatePublicHTTP(context.Background(), tc.raw, r); (err == nil) != tc.ok {
			t.Fatalf("ValidatePublicHTTP(%q) error=%v, ok=%v", tc.raw, err, tc.ok)
		}
	}
}

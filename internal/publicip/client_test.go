package publicip

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
)

func TestIpifyClient_Resolve(t *testing.T) {
	tests := []struct {
		name string
		body string
		want netip.Addr
	}{
		{
			// TEST-NET-1: 192.0.2.0/24
			name: "TEST-NET-1",
			body: "192.0.2.10",
			want: netip.MustParseAddr("192.0.2.10"),
		},
		{
			// TEST-NET-2: 198.51.100.0/24
			name: "TEST-NET-2",
			body: "198.51.100.10",
			want: netip.MustParseAddr("198.51.100.10"),
		},
		{
			// TEST-NET-3: 203.0.113.0/24
			name: "TEST-NET-3",
			body: "203.0.113.10",
			want: netip.MustParseAddr("203.0.113.10"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = fmt.Fprint(w, tt.body)
			}))
			defer server.Close()

			client := IpifyClient{endpoint: server.URL}
			got, err := client.Resolve(context.Background(), 1)
			if err != nil {
				t.Fatalf("Ipify Resolve() Error: %v", err)
			}

			if got != tt.want {
				t.Fatalf("Ipify Resolve(): got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIpifyClient_Resolve_InvalidIP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "not-an-ip")
	}))
	defer server.Close()

	client := IpifyClient{endpoint: server.URL}
	_, err := client.Resolve(context.Background(), 1)

	if err == nil {
		t.Fatalf("Ipify Resolve(): got nil, want error")
	}
}

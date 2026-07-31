package publicip

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
	"time"
)

func TestPlainClient_Resolve(t *testing.T) {
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

			client := PlainClient{endpoint: server.URL}
			got, err := client.Resolve(context.Background(), time.Second)
			if err != nil {
				t.Fatalf("Ipify Resolve() Error: %v", err)
			}

			if got != tt.want {
				t.Fatalf("Ipify Resolve(): got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlainClient_Resolve_InvalidIP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "not-an-ip")
	}))
	defer server.Close()

	client := PlainClient{endpoint: server.URL}
	_, err := client.Resolve(context.Background(), time.Second)

	if err == nil {
		t.Fatalf("Plain Resolve(): got nil, want error")
	}
}

func TestJSONClient_Resolve(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"ip":"203.0.113.10"}`)
	}))
	defer server.Close()

	client := JSONClient{endpoint: server.URL, jsonPath: "$.ip"}
	got, err := client.Resolve(context.Background(), time.Second)
	if err != nil {
		t.Fatalf("JSON Resolve() error = %v", err)
	}

	want := netip.MustParseAddr("203.0.113.10")
	if got != want {
		t.Fatalf("JSON Resolve(): got %v, want %v", got, want)
	}
}

func TestJSONClient_Resolve_Errors(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		jsonPath string
	}{
		{
			name:     "missing path",
			body:     `{"ip":"203.0.113.10"}`,
			jsonPath: "$.address",
		},
		{
			name:     "non-string value",
			body:     `{"ip":203}`,
			jsonPath: "$.ip",
		},
		{
			name:     "invalid ip",
			body:     `{"ip":"not-an-ip"}`,
			jsonPath: "$.ip",
		},
		{
			name:     "invalid json",
			body:     `not-json`,
			jsonPath: "$.ip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = fmt.Fprint(w, tt.body)
			}))
			defer server.Close()

			client := JSONClient{endpoint: server.URL, jsonPath: tt.jsonPath}
			_, err := client.Resolve(context.Background(), time.Second)
			if err == nil {
				t.Fatal("JSON Resolve() error = nil, want error")
			}
		})
	}
}

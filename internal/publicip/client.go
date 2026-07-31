package publicip

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"github.com/fudoge/cf-ddns-client/internal/config"
	"github.com/ohler55/ojg/jp"
)

type Options struct {
	Endpoint     string
	ResponseType config.PublicIPResponseType
	JSONPath     string
	Timeout      time.Duration
}

type IPResolver interface {
	Resolve(ctx context.Context, timeout time.Duration) (netip.Addr, error)
}

type PlainClient struct {
	endpoint string
}

func NewPlainClient(endpoint string) *PlainClient {
	return &PlainClient{
		endpoint: endpoint,
	}
}

func (c PlainClient) Resolve(ctx context.Context, timeout time.Duration) (netip.Addr, error) {
	client := &http.Client{}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint, nil)
	if err != nil {
		return netip.Addr{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return netip.Addr{}, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return netip.Addr{}, fmt.Errorf("error fetching public ip: http response %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return netip.Addr{}, err
	}

	trimmed := strings.TrimSpace(string(body))
	ip, err := netip.ParseAddr(trimmed)
	if err != nil {
		return netip.Addr{}, err
	}

	return ip, nil
}

type JSONClient struct {
	endpoint string
	jsonPath string
}

func NewJSONClient(endpoint, jsonpath string) *JSONClient {
	return &JSONClient{
		endpoint: endpoint,
		jsonPath: jsonpath,
	}
}

func (c JSONClient) Resolve(ctx context.Context, timeout time.Duration) (netip.Addr, error) {
	client := &http.Client{}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint, nil)
	if err != nil {
		return netip.Addr{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return netip.Addr{}, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return netip.Addr{}, fmt.Errorf("error fetching public ip: http response %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return netip.Addr{}, err
	}

	var data any
	err = json.Unmarshal([]byte(body), &data)
	if err != nil {
		return netip.Addr{}, fmt.Errorf("error parsing JSON: %v", err)
	}

	path, err := jp.ParseString(c.jsonPath)
	if err != nil {
		return netip.Addr{}, fmt.Errorf("error parsing jsonpath: %v", err)
	}

	got := path.Get(data)
	if len(got) != 1 {
		return netip.Addr{}, fmt.Errorf("jsonpath query failed: got %d results, want 1", len(got))
	}

	value := got[0]
	addr, ok := value.(string)
	if !ok {
		return netip.Addr{}, fmt.Errorf("failed to fetch data: %v", err)
	}

	trimmed := strings.TrimSpace(addr)
	ip, err := netip.ParseAddr(trimmed)
	if err != nil {
		return netip.Addr{}, err
	}

	return ip, nil
}

func GetIP(ctx context.Context, options *Options) (netip.Addr, error) {
	if options.ResponseType == config.ResponseTypePlain {
		log.Printf("fetching public IP from %s", options.Endpoint)
		client := NewPlainClient(options.Endpoint)
		ip, err := client.Resolve(ctx, options.Timeout)

		return ip, err
	} else {
		log.Printf("fetching public IP from %s, with jsonpath %s", options.Endpoint, options.JSONPath)
		client := NewJSONClient(options.Endpoint, options.JSONPath)
		ip, err := client.Resolve(ctx, options.Timeout)

		return ip, err
	}

}

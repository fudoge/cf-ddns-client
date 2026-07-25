package publicip

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/netip"
	"time"
)

type IPResolver interface {
	Resolve(ctx context.Context, timeoutsec int) (netip.Addr, error)
}

type IpifyClient struct {
	endpoint string
}

func NewIpifyClient() *IpifyClient {
	return &IpifyClient{
		endpoint: "https://api.ipify.org",
	}
}

func (c IpifyClient) Resolve(ctx context.Context, timeoutsec int) (netip.Addr, error) {
	client := &http.Client{}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutsec)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint, nil)
	if err != nil {
		return netip.Addr{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return netip.Addr{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return netip.Addr{}, err
	}

	ip, err := netip.ParseAddr(string(body))
	if err != nil {
		return netip.Addr{}, err
	}

	return ip, nil
}

func GetIP(ctx context.Context, timeoutsec int) (netip.Addr, error) {
	var ip netip.Addr

	log.Printf("fetching public IP from ipify")
	ipify := NewIpifyClient()
	ip, err := ipify.Resolve(ctx, timeoutsec)
	if err == nil {
		return ip, nil
	}

	return netip.Addr{}, fmt.Errorf("failed to get public IP from ipify: %w", err)
}

package cloudflare

import (
	"context"
	"log"
	"net/netip"
	"time"

	"github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/dns"
	"github.com/cloudflare/cloudflare-go/v7/option"
)

type Client struct {
	client  *cloudflare.Client
	zoneID  string
	timeout time.Duration
}

type ARecord struct {
	ID      string
	Name    string
	Content netip.Addr
}

func NewClient(token, zoneID string, timeout time.Duration) *Client {
	return &Client{
		client: cloudflare.NewClient(
			option.WithAPIToken(token),
		),
		zoneID:  zoneID,
		timeout: timeout,
	}
}

func (c *Client) ListARecords(ctx context.Context, name string) ([]ARecord, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	page, err := c.client.DNS.Records.List(ctx, dns.RecordListParams{
		ZoneID: cloudflare.F(c.zoneID),
		Name:   cloudflare.F(dns.RecordListParamsName{Exact: cloudflare.F(name)}),
		Type:   cloudflare.F(dns.RecordListParamsTypeA),
	})
	if err != nil {
		return nil, err
	}
	resp := page.Result

	records := make([]ARecord, 0)
	for i := range resp {
		content, err := netip.ParseAddr(resp[i].Content)
		if err != nil {
			log.Printf("skipping A record with invalid IP address %q: %v", resp[i].Content, err)
			continue
		}

		rec := ARecord{
			ID:      resp[i].ID,
			Name:    resp[i].Name,
			Content: content,
		}
		records = append(records, rec)

	}

	return records, nil
}

func (c *Client) CreateARecord(ctx context.Context, name string, ip netip.Addr) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	_, err := c.client.DNS.Records.New(ctx, dns.RecordNewParams{
		ZoneID: cloudflare.F(c.zoneID),
		Body: dns.ARecordParam{
			Name: cloudflare.F(name),
			// TODO: replace TTL1 with a config value.
			TTL:     cloudflare.F(dns.TTL1),
			Type:    cloudflare.F(dns.ARecordTypeA),
			Content: cloudflare.F(ip.String()),
		},
	})

	return err
}

func (c *Client) DeleteARecord(ctx context.Context, recordID string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	_, err := c.client.DNS.Records.Delete(ctx, recordID, dns.RecordDeleteParams{
		ZoneID: cloudflare.F(c.zoneID),
	})

	return err
}

package ddns

import (
	"cf-ddns-client/internal/cloudflare"
	"context"
	"net/netip"
)

type DNSClient interface {
	ListARecords(ctx context.Context, name string) ([]cloudflare.ARecord, error)
	CreateARecord(ctx context.Context, name string, ip netip.Addr) error
	DeleteARecord(ctx context.Context, recordID string) error
}

type Syncer struct {
	dns DNSClient
}

func NewSyncer(c DNSClient) *Syncer {
	return &Syncer{dns: c}
}

func (s *Syncer) Append(ctx context.Context, name string, ip netip.Addr) error {
	records, err := s.dns.ListARecords(ctx, name)
	if err != nil {
		return err
	}

	var exists bool
	for _, record := range records {
		if record.Content == ip {
			exists = true
			break
		}
	}

	if !exists {
		err := s.dns.CreateARecord(ctx, name, ip)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Syncer) Replace(ctx context.Context, name string, ip netip.Addr) error {
	records, err := s.dns.ListARecords(ctx, name)
	if err != nil {
		return err
	}

	var exists bool
	for _, record := range records {
		if record.Content == ip {
			exists = true
			continue
		}
		if err := s.dns.DeleteARecord(ctx, record.ID); err != nil {
			return err
		}
	}

	if !exists {
		err := s.dns.CreateARecord(ctx, name, ip)
		if err != nil {
			return err
		}
	}

	return nil
}

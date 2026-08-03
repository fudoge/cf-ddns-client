package ddns

import (
	"context"
	"net/netip"

	"github.com/cloudflare/cloudflare-go/v7/dns"
	"github.com/fudoge/cf-ddns-client/internal/cloudflare"
)

type DNSClient interface {
	ListARecords(ctx context.Context, name string) ([]cloudflare.ARecord, error)
	CreateARecord(ctx context.Context, name string, ip netip.Addr, ttl dns.TTL) error
	UpdateARecord(ctx context.Context, recordID, name string, ip netip.Addr, ttl dns.TTL) error
	DeleteARecord(ctx context.Context, recordID string) error
}

type Syncer struct {
	dns DNSClient
}

func NewSyncer(c DNSClient) *Syncer {
	return &Syncer{dns: c}
}

func (s *Syncer) Append(ctx context.Context, name string, ip netip.Addr, ttl dns.TTL) error {
	records, err := s.dns.ListARecords(ctx, name)
	if err != nil {
		return err
	}

	var exists bool
	for _, record := range records {
		if record.Content == ip {
			exists = true
			if record.TTL != ttl {
				err := s.dns.UpdateARecord(ctx, record.ID, name, ip, ttl)
				if err != nil {
					return err
				}
			}
			break
		}
	}

	if !exists {
		err := s.dns.CreateARecord(ctx, name, ip, ttl)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Syncer) Replace(ctx context.Context, name string, ip netip.Addr, ttl dns.TTL) error {
	records, err := s.dns.ListARecords(ctx, name)
	if err != nil {
		return err
	}

	var found bool
	for _, record := range records {
		if record.Content != ip {
			err := s.dns.DeleteARecord(ctx, record.ID)
			if err != nil {
				return err
			}
			continue
		}

		if !found {
			found = true
			if record.TTL != ttl {
				err := s.dns.UpdateARecord(ctx, record.ID, name, ip, ttl)
				if err != nil {
					return err
				}
			}
			continue
		}

		if err := s.dns.DeleteARecord(ctx, record.ID); err != nil {
			return err
		}
	}

	if !found {
		err := s.dns.CreateARecord(ctx, name, ip, ttl)
		if err != nil {
			return err
		}
	}

	return nil
}

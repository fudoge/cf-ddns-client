package ddns

import (
	"context"
	"net/netip"
	"reflect"
	"testing"

	"github.com/cloudflare/cloudflare-go/v7/dns"
	"github.com/fudoge/cf-ddns-client/internal/cloudflare"
)

type fakeDNSClient struct {
	records []cloudflare.ARecord

	created []fakeRecordChange
	updated []fakeRecordChange
	deleted []string

	listErr   error
	createErr error
	updateErr error
	deleteErr error
}

type fakeRecordChange struct {
	recordID string
	name     string
	ip       netip.Addr
	ttl      dns.TTL
}

func (f *fakeDNSClient) ListARecords(ctx context.Context, name string) ([]cloudflare.ARecord, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.records, nil
}

func (f *fakeDNSClient) CreateARecord(ctx context.Context, name string, ip netip.Addr, ttl dns.TTL) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.created = append(f.created, fakeRecordChange{
		name: name,
		ip:   ip,
		ttl:  ttl,
	})
	return nil
}

func (f *fakeDNSClient) UpdateARecord(ctx context.Context, recordID, name string, ip netip.Addr, ttl dns.TTL) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	f.updated = append(f.updated, fakeRecordChange{
		recordID: recordID,
		name:     name,
		ip:       ip,
		ttl:      ttl,
	})
	return nil
}

func (f *fakeDNSClient) DeleteARecord(ctx context.Context, recordID string) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	f.deleted = append(f.deleted, recordID)
	return nil
}

func TestSyncer_Append(t *testing.T) {
	ip := netip.MustParseAddr("203.0.113.10")
	name := "home.example.com"

	tests := []struct {
		name        string
		records     []cloudflare.ARecord
		wantCreated int
	}{
		{
			name:        "creates record when IP is missing",
			records:     []cloudflare.ARecord{},
			wantCreated: 1,
		},
		{
			name: "does not create record when IP exists",
			records: []cloudflare.ARecord{
				{
					ID:      "record-1",
					Name:    name,
					Content: ip,
					TTL:     dns.TTL1,
				},
			},
			wantCreated: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &fakeDNSClient{
				records: tt.records,
			}
			syncer := NewSyncer(client)

			err := syncer.Append(context.Background(), name, ip, dns.TTL1)
			if err != nil {
				t.Fatalf("Append() error = %v", err)
			}

			if len(client.created) != tt.wantCreated {
				t.Fatalf("created records = %d, want %d", len(client.created), tt.wantCreated)
			}

			if tt.wantCreated == 1 && client.created[0].ip != ip {
				t.Fatalf("created IP = %v, want %v", client.created[0].ip, ip)
			}
		})
	}
}

func TestSyncer_Replace(t *testing.T) {
	ip := netip.MustParseAddr("203.0.113.10")
	name := "home.example.com"

	tests := []struct {
		name        string
		records     []cloudflare.ARecord
		ttl         dns.TTL
		wantCreated int
		wantUpdated []fakeRecordChange
		wantDeleted []string
	}{
		{
			name: "keeps current IP record when TTL matches",
			ttl:  dns.TTL1,
			records: []cloudflare.ARecord{
				{
					ID:      "record-1",
					Name:    name,
					Content: ip,
					TTL:     dns.TTL1,
				},
			},
			wantCreated: 0,
			wantDeleted: nil,
		},
		{
			name: "updates current IP record when only TTL differs",
			ttl:  dns.TTL(300),
			records: []cloudflare.ARecord{
				{
					ID:      "record-1",
					Name:    name,
					Content: ip,
					TTL:     dns.TTL1,
				},
			},
			wantCreated: 0,
			wantUpdated: []fakeRecordChange{
				{
					recordID: "record-1",
					name:     name,
					ip:       ip,
					ttl:      dns.TTL(300),
				},
			},
			wantDeleted: nil,
		},
		{
			name: "deletes old records and creates current IP if missing",
			ttl:  dns.TTL(300),
			records: []cloudflare.ARecord{
				{
					ID:      "record-1",
					Name:    name,
					Content: netip.MustParseAddr("198.51.100.10"),
				},
				{
					ID:      "record-2",
					Name:    name,
					Content: netip.MustParseAddr("192.0.2.10"),
				},
			},
			wantCreated: 1,
			wantDeleted: []string{"record-1", "record-2"},
		},
		{
			name: "updates first current IP record and deletes duplicates and old records",
			ttl:  dns.TTL(600),
			records: []cloudflare.ARecord{
				{
					ID:      "record-1",
					Name:    name,
					Content: netip.MustParseAddr("198.51.100.10"),
					TTL:     dns.TTL1,
				},
				{
					ID:      "record-2",
					Name:    name,
					Content: ip,
					TTL:     dns.TTL1,
				},
				{
					ID:      "record-3",
					Name:    name,
					Content: ip,
					TTL:     dns.TTL(600),
				},
			},
			wantCreated: 0,
			wantUpdated: []fakeRecordChange{
				{
					recordID: "record-2",
					name:     name,
					ip:       ip,
					ttl:      dns.TTL(600),
				},
			},
			wantDeleted: []string{"record-1", "record-3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &fakeDNSClient{
				records: tt.records,
			}
			syncer := NewSyncer(client)

			err := syncer.Replace(context.Background(), name, ip, tt.ttl)
			if err != nil {
				t.Fatalf("Replace() error = %v", err)
			}

			if len(client.created) != tt.wantCreated {
				t.Fatalf("created records = %d, want %d", len(client.created), tt.wantCreated)
			}

			if !reflect.DeepEqual(client.updated, tt.wantUpdated) {
				t.Fatalf("updated records = %v, want %v", client.updated, tt.wantUpdated)
			}

			if !reflect.DeepEqual(client.deleted, tt.wantDeleted) {
				t.Fatalf("deleted records = %v, want %v", client.deleted, tt.wantDeleted)
			}

			if tt.wantCreated == 1 {
				if client.created[0].ip != ip {
					t.Fatalf("created IP = %v, want %v", client.created[0].ip, ip)
				}
				if client.created[0].ttl != tt.ttl {
					t.Fatalf("created TTL = %v, want %v", client.created[0].ttl, tt.ttl)
				}
			}
		})
	}
}

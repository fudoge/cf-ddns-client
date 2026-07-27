package ddns

import (
	"cf-ddns-client/internal/cloudflare"
	"context"
	"net/netip"
	"reflect"
	"testing"
)

type fakeDNSClient struct {
	records []cloudflare.ARecord

	created []netip.Addr
	deleted []string

	listErr   error
	createErr error
	deleteErr error
}

func (f *fakeDNSClient) ListARecords(ctx context.Context, name string) ([]cloudflare.ARecord, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.records, nil
}

func (f *fakeDNSClient) CreateARecord(ctx context.Context, name string, ip netip.Addr) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.created = append(f.created, ip)
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

			err := syncer.Append(context.Background(), name, ip)
			if err != nil {
				t.Fatalf("Append() error = %v", err)
			}

			if len(client.created) != tt.wantCreated {
				t.Fatalf("created records = %d, want %d", len(client.created), tt.wantCreated)
			}

			if tt.wantCreated == 1 && client.created[0] != ip {
				t.Fatalf("created IP = %v, want %v", client.created[0], ip)
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
		wantCreated int
		wantDeleted []string
	}{
		{
			name: "keeps current IP record if exists",
			records: []cloudflare.ARecord{
				{
					ID:      "record-1",
					Name:    name,
					Content: ip,
				},
			},
			wantCreated: 0,
			wantDeleted: nil,
		},
		{
			name: "deletes old records and creates current IP if missing",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &fakeDNSClient{
				records: tt.records,
			}
			syncer := NewSyncer(client)

			err := syncer.Replace(context.Background(), name, ip)
			if err != nil {
				t.Fatalf("Replace() error = %v", err)
			}

			if len(client.created) != tt.wantCreated {
				t.Fatalf("created records = %d, want %d", len(client.created), tt.wantCreated)
			}

			if !reflect.DeepEqual(client.deleted, tt.wantDeleted) {
				t.Fatalf("deleted records = %v, want %v", client.deleted, tt.wantDeleted)
			}

			if tt.wantCreated == 1 && client.created[0] != ip {
				t.Fatalf("created IP = %v, want %v", client.created[0], ip)
			}
		})
	}
}

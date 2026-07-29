package main

import (
	"context"
	"log"

	"github.com/fudoge/cf-ddns-client/internal/cloudflare"
	"github.com/fudoge/cf-ddns-client/internal/config"
	"github.com/fudoge/cf-ddns-client/internal/ddns"
	"github.com/fudoge/cf-ddns-client/internal/publicip"
)

func main() {
	ctx := context.Background()

	config, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ip, err := publicip.GetIP(ctx, config.Timeout)
	if err != nil {
		log.Fatalf("failed to fetch public IP: %v", err)
	}

	client := cloudflare.NewClient(config.CFConfig.APIToken, config.CFConfig.ZoneID, config.Timeout)
	syncer := ddns.NewSyncer(client)

	switch config.Mode {
	case "append":
		err = syncer.Append(ctx, config.Name, ip)
	case "replace":
		err = syncer.Replace(ctx, config.Name, ip)
	default:
		log.Fatalf("unsupported mode %q", config.Mode)
	}
	if err != nil {
		log.Fatalf("failed to sync DNS record: %v", err)
	}
}

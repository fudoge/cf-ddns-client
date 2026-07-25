package main

import (
	"cf-ddns-client/internal/cloudflare"
	"cf-ddns-client/internal/config"
	"cf-ddns-client/internal/ddns"
	"cf-ddns-client/internal/publicip"
	"context"
	"log"
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

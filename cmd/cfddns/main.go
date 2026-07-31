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

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	options := &publicip.Options{
		Endpoint:     cfg.PublicIPConfig.Endpoint,
		ResponseType: cfg.PublicIPConfig.ResponseType,
		JSONPath:     cfg.PublicIPConfig.JSONPath,
		Timeout:      cfg.Timeout,
	}
	ip, err := publicip.GetIP(ctx, options)
	if err != nil {
		log.Fatalf("failed to fetch public IP: %v", err)
	}

	client := cloudflare.NewClient(cfg.CFConfig.APIToken, cfg.CFConfig.ZoneID, cfg.Timeout)
	syncer := ddns.NewSyncer(client)

	switch cfg.Mode {
	case config.ModeAppend:
		err = syncer.Append(ctx, cfg.Name, ip, cfg.TTL)
	case config.ModeReplace:
		err = syncer.Replace(ctx, cfg.Name, ip, cfg.TTL)
	default:
		log.Fatalf("unsupported mode %q", cfg.Mode)
	}
	if err != nil {
		log.Fatalf("failed to sync DNS record: %v", err)
	}
}

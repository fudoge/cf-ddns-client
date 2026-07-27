package config

import (
	"flag"
	"fmt"
	"os"
	"slices"
)

type Config struct {
	Name    string
	Timeout int

	Mode     string
	CFConfig *CloudflareConfig
}

type flagVars struct {
	Timeout int
	Mode    string
}

type CloudflareConfig struct {
	ZoneID   string
	APIToken string
}

var allowedModes []string = []string{"replace", "append"}

func Load() (*Config, error) {

	name, err := requireEnv("DOMAIN_NAME")
	if err != nil {
		return nil, err
	}

	zoneID, err := requireEnv("ZONE_ID")
	if err != nil {
		return nil, err
	}

	apiToken, err := requireEnv("CF_API_TOKEN")
	if err != nil {
		return nil, err
	}

	flagvars, err := parseFlags(os.Args[1:])
	if err != nil {
		return nil, err
	}

	c := &Config{
		Name:    name,
		Timeout: flagvars.Timeout,

		Mode: flagvars.Mode,
		CFConfig: &CloudflareConfig{
			ZoneID:   zoneID,
			APIToken: apiToken,
		},
	}

	return c, nil
}

func requireEnv(key string) (string, error) {
	val, exists := os.LookupEnv(key)
	if !exists {
		return "", fmt.Errorf("missing required environment variable %s", key)
	}

	return val, nil
}

func parseFlags(args []string) (*flagVars, error) {
	flags := flag.NewFlagSet("cfddns", flag.ContinueOnError)
	timeout := flags.Int("timeout", 2, "Request timeout in seconds")
	mode := flags.String("mode", "replace",
		"DNS sync mode: replace keeps only the current public IP; append adds it if missing")
	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	if *timeout <= 0 {
		return nil, fmt.Errorf("timeout must be greater than zero")
	}

	if !slices.Contains(allowedModes, *mode) {
		return nil, fmt.Errorf("unsupported mode %q; allowed modes: %v", *mode, allowedModes)
	}

	return &flagVars{
		Timeout: *timeout,
		Mode:    *mode,
	}, nil
}

package config

import (
	"flag"
	"fmt"
	"os"
	"slices"
	"time"
)

type UpdateMode int

const (
	ModeReplace UpdateMode = iota
	ModeAppend
)

type PublicIPResponseType int

const (
	ResponseTypePlain PublicIPResponseType = iota
	ResponseTypeJSON
)

type Config struct {
	Name    string
	Timeout time.Duration

	Mode           UpdateMode
	CFConfig       *CloudflareConfig
	PublicIPConfig *PublicIPSourceConfig
}

type flagVars struct {
	Timeout  int
	Mode     string
	Endpoint string
	JSONPath string
}

type CloudflareConfig struct {
	ZoneID   string
	APIToken string
}

type PublicIPSourceConfig struct {
	Endpoint     string
	ResponseType PublicIPResponseType
	JSONPath     string
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

	updateMode := ModeReplace
	if flagvars.Mode == "append" {
		updateMode = ModeAppend
	}

	responseType := ResponseTypePlain
	if flagvars.JSONPath != "" {
		responseType = ResponseTypeJSON
	}

	c := &Config{
		Name:    name,
		Timeout: time.Duration(flagvars.Timeout) * time.Second,

		Mode: updateMode,
		CFConfig: &CloudflareConfig{
			ZoneID:   zoneID,
			APIToken: apiToken,
		},
		PublicIPConfig: &PublicIPSourceConfig{
			Endpoint:     flagvars.Endpoint,
			ResponseType: responseType,
			JSONPath:     flagvars.JSONPath,
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
	endpoint := flags.String("endpoint", "https://api.ipify.org", "Public IP provider URL endpoint. \nDefault: https://api.ipify.org")
	jsonPath := flags.String("jsonpath", "", "Public IP API Response JSON path")
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
		Timeout:  *timeout,
		Mode:     *mode,
		Endpoint: *endpoint,
		JSONPath: *jsonPath,
	}, nil
}

# Cloudflare DDNS Client

Small DDNS client for Cloudflare DNS A records.

The client fetches the current public IPv4 address from a configurable endpoint,
then syncs the configured Cloudflare DNS record in either `replace` or `append`
mode. The default endpoint is ipify.

Cloudflare Go SDK docs:
[https://developers.cloudflare.com/api/go/resources/dns/](https://developers.cloudflare.com/api/go/resources/dns/)

## Configuration

Required environment variables:

```text
CF_API_TOKEN  Cloudflare API token with DNS read/write access
ZONE_ID       Cloudflare zone ID
DOMAIN_NAME   DNS record name to sync, e.g. home.example.com
```

Flags:

```text
--mode       DNS sync mode. Allowed: replace, append. Default: replace
--timeout    Per-request timeout in seconds. Default: 2
--ttl        DNS record TTL in seconds. Must be an integer.
             Allowed: 1 for automatic, or 60-86400. Default: 1
--endpoint   Public IP provider endpoint. Default: https://api.ipify.org
--jsonpath   JSONPath expression for JSON responses. If set, the endpoint
             response is parsed as JSON and the selected value is used as the IP.
```

Public IP provider examples:

```bash
# Plain text response, compatible with the default ipify endpoint.
cfddns --endpoint https://api.ipify.org

# JSON response.
cfddns --endpoint 'https://api64.ipify.org?format=json' --jsonpath '$.ip'
```

Modes:

```text
replace
  Keep only the current public IPv4 address for DOMAIN_NAME.
  Existing A records with other IPs are deleted.
  If the current IP already exists but has a different TTL, it is updated.

append
  Add the current public IPv4 address if it is missing.
  Existing A records are kept.
```

## Container

Published images are available from GitHub Container Registry:

```text
ghcr.io/fudoge/cf-ddns-client
```

Run once with Docker:

```bash
docker run --rm \
  -e CF_API_TOKEN=xxxx \
  -e ZONE_ID=xxxx \
  -e DOMAIN_NAME=home.example.com \
  ghcr.io/fudoge/cf-ddns-client:latest \
  --mode replace \
  --timeout 2 \
  --ttl 1 \
  --endpoint https://api.ipify.org
```

For a tagged release, pin the image tag instead of using `latest`:

```bash
docker run --rm \
  -e CF_API_TOKEN=xxxx \
  -e ZONE_ID=xxxx \
  -e DOMAIN_NAME=home.example.com \
  ghcr.io/fudoge/cf-ddns-client:0.1.0 \
  --mode replace \
  --timeout 2 \
  --ttl 1 \
  --endpoint https://api.ipify.org
```

## Kubernetes

CronJob example:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: cfddns
type: Opaque
stringData:
  CF_API_TOKEN: xxxx
  ZONE_ID: xxxx
---
apiVersion: batch/v1
kind: CronJob
metadata:
  name: cfddns
spec:
  schedule: "*/5 * * * *"
  concurrencyPolicy: Forbid
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 3
  jobTemplate:
    spec:
      template:
        spec:
          restartPolicy: OnFailure
          containers:
            - name: cfddns
              image: ghcr.io/fudoge/cf-ddns-client:0.1.0
              imagePullPolicy: IfNotPresent
              args:
                - --mode
                - replace
                - --timeout
                - "2"
                - --ttl
                - "1"
              env:
                - name: DOMAIN_NAME
                  value: home.example.com
                - name: CF_API_TOKEN
                  valueFrom:
                    secretKeyRef:
                      name: cfddns
                      key: CF_API_TOKEN
                - name: ZONE_ID
                  valueFrom:
                    secretKeyRef:
                      name: cfddns
                      key: ZONE_ID
```

## Binary

Download a release archive from the GitHub Releases page, or build from source:

```bash
go build -o cfddns ./cmd/cfddns
chmod +x cfddns
sudo mv cfddns /usr/local/bin/cfddns
```

Run the binary:

```bash
CF_API_TOKEN=xxxx \
ZONE_ID=xxxx \
DOMAIN_NAME=home.example.com \
cfddns --mode replace --timeout 2 --ttl 300
```

## Systemd Setup

```ini
# /etc/systemd/system/cfddns.service
[Unit]
Description=Cloudflare DDNS Client
After=network-online.target
Wants=network-online.target

[Service]
Type=oneshot
ExecStart=/usr/local/bin/cfddns --mode replace --timeout 2 --ttl 300
Environment=CF_API_TOKEN=xxxx
Environment=ZONE_ID=xxxx
Environment=DOMAIN_NAME=home.example.com

NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true

[Install]
WantedBy=multi-user.target
```

```ini
# /etc/systemd/system/cfddns.timer
[Unit]
Description=Run Cloudflare DDNS Client periodically

[Timer]
OnBootSec=30
OnUnitActiveSec=5min
Persistent=true

[Install]
WantedBy=timers.target
```

Apply:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now cfddns.timer
```

Check:

```bash
systemctl list-timers | grep cfddns
journalctl -u cfddns.service
```

## LaunchAgent Setup

Create `~/Library/LaunchAgents/com.yourname.cfddns.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.yourname.cfddns</string>

    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/cfddns</string>
        <string>--mode</string>
        <string>replace</string>
        <string>--timeout</string>
        <string>2</string>
        <string>--ttl</string>
        <string>300</string>
    </array>

    <key>EnvironmentVariables</key>
    <dict>
        <key>CF_API_TOKEN</key>
        <string>YOUR_CF_API_TOKEN</string>
        <key>ZONE_ID</key>
        <string>YOUR_CF_ZONE_ID</string>
        <key>DOMAIN_NAME</key>
        <string>home.example.com</string>
    </dict>

    <key>RunAtLoad</key>
    <true/>

    <key>StartInterval</key>
    <integer>300</integer>

    <key>StandardOutPath</key>
    <string>/usr/local/var/log/cfddns.log</string>
    <key>StandardErrorPath</key>
    <string>/usr/local/var/log/cfddns.err</string>
</dict>
</plist>
```

Apply:

```bash
launchctl load ~/Library/LaunchAgents/com.yourname.cfddns.plist
launchctl start com.yourname.cfddns
```

Check:

```bash
launchctl list | grep cfddns
```

To rotate logs on macOS, create `/etc/newsyslog.d/cfddns.conf`:

```conf
# logfilename                    mode    count   size    when    flags
/usr/local/var/log/cfddns.log    644     5       1000    *       J
/usr/local/var/log/cfddns.err    644     5       1000    *       J
```

## License

MIT. See [LICENSE](LICENSE).

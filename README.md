# Synthient CLI

Official command-line access to Synthient account status, IP and domain intelligence, live feed streams, Parquet feed snapshots, and protobuf schemas exposed through gRPC reflection.

The CLI is built for developers and automation:

- Human-readable terminal output by default, with a polished color palette for interactive use.
- `json` and `csv` output for scripts, CI jobs, data pipelines, and notebooks.
- Production Synthient endpoints by default. No config file is required for normal usage.
- Optional config profiles for custom endpoint routing.
- Live NDJSON feed streaming with filters, reconnects, duration limits, and event limits.
- A built-in Model Context Protocol (MCP) server that exposes Synthient lookups, feeds, and schemas as tools for MCP-compatible clients.

## Install

Install from Homebrew:

```bash
brew install synthient/tap/synthient
```

Install with Go:

```bash
go install github.com/synthient/cli/cmd@latest
```

Build from a local checkout:

```bash
git clone https://github.com/synthient/cli.git
cd cli
go build -o ./dist/synthient ./cmd
./dist/synthient --version
```

## Authentication

For automation, pass the API key through the environment:

```bash
export SYNTHIENT_API_KEY="..."
synthient status
```

For interactive local use, store the key in the OS keychain:

```bash
synthient auth
```

Remove the keychain credential:

```bash
synthient auth --logout
```

Credential lookup order:

```txt
1. SYNTHIENT_API_KEY
2. .env in the current working directory
3. OS keychain
```

`.env` example:

```bash
SYNTHIENT_API_KEY="..."
```

Use `synthient status` to confirm which auth source is active without printing the secret.

## Demo Gallery

| Status | Lookup |
| --- | --- |
| ![Status and auth demo](demos/auth.gif) | ![Lookup demo](demos/lookup.gif) |

| Feeds | Streaming |
| --- | --- |
| ![Feed snapshot demo](demos/feeds.gif) | ![Proxy stream demo](demos/stream.gif) |

| Download | gRPC Schemas |
| --- | --- |
| ![Snapshot download demo](demos/download.gif) | ![gRPC schema demo](demos/grpc.gif) |

## Quick Start

Check configuration, authentication, quota, and scopes:

```bash
synthient status
synthient account
synthient scopes
```

Look up an IP:

```bash
synthient lookup 8.8.8.8
```

Look up a batch of IPs:

```bash
synthient lookup 8.8.8.8 1.1.1.1 --format csv
```

Read IPs from stdin:

```bash
printf "8.8.8.8\n1.1.1.1\n" | synthient lookup --format json
```

Look up Helios domain intelligence:

```bash
synthient lookup domain example.com
```

List feed streams and snapshots:

```bash
synthient feeds streams
synthient feeds snapshots proxies --limit 10
```

Resolve the newest listed proxy snapshot and inspect it:

```bash
SNAPSHOT=$(synthient feeds snapshots proxies --limit 1 --format json | jq -r '.feeds[0].id')
synthient feeds meta proxies "$SNAPSHOT"
synthient feeds schema proxies "$SNAPSHOT"
synthient feeds checksum proxies "$SNAPSHOT"
```

Download a Parquet snapshot with progress:

```bash
SNAPSHOT=$(synthient feeds snapshots proxies --limit 1 --format json | jq -r '.feeds[0].id')
synthient feeds download proxies "$SNAPSHOT" proxies-latest.parquet --verify --force
```

Stream live residential proxy observations for five seconds:

```bash
synthient stream proxies --filter type=RESIDENTIAL_PROXY --duration 5s
```

Export protobuf schema information through gRPC reflection:

```bash
synthient grpc schema synthient.v1.SynthientService
```

## Command Map

```txt
synthient
  auth                 Store or remove an API key in the OS keychain
  status               Show CLI config, auth source, endpoint, quota, and scope status
  account              Show account, scope, and quota details
  scopes               Show known Synthient scopes and whether the active key has them
  lookup               Lookup IP intelligence; accepts one IP, many IPs, or stdin
  lookup domain        Lookup domain intelligence from Helios observations
  feeds streams        List supported feed streams
  feeds snapshots      List available daily and hourly Parquet snapshots
  feeds meta           Show snapshot metadata plus Parquet schema
  feeds schema         Show only the Parquet schema for a snapshot
  feeds checksum       Show the expected SHA-256 checksum for a snapshot
  feeds download       Download an explicit snapshot to a Parquet file
  download             Convenience wrapper for snapshot downloads
  stream               Stream live NDJSON feed events
  grpc schema          Output protobuf descriptors using gRPC reflection
  mcp                  Run a Model Context Protocol server over stdio
```

## Output Modes

Most finite read commands support:

```txt
--format text|json|csv
--output -|file
```

Examples:

```bash
synthient status --format json
synthient account --format csv --output account.csv
synthient lookup 8.8.8.8 --format json | jq .
synthient feeds snapshots proxies --format csv --output snapshots.csv
```

`text` is optimized for humans. `json` and `csv` are intended for automation.

`synthient stream` intentionally emits NDJSON and does not take `--format`. Use `--pretty` only for local inspection:

```bash
synthient stream proxies --max-events 3 --pretty
synthient stream proxies --max-events 250 --output proxies.ndjson
```

Disable color for logs, snapshots, or terminals that do not support ANSI:

```bash
synthient --no-color status
NO_COLOR=1 synthient status
```

Suppress human-oriented progress output:

```bash
synthient --quiet feeds download proxies "$SNAPSHOT" proxies.parquet --force
```

## Configuration

No config file is required for production usage. By default, the CLI uses the production Synthient API, feed, and gRPC endpoints from the Synthient SDK.

Use a config file only when you need custom endpoints or named profiles:

```txt
~/.config/synthient/config.toml
```

Example:

```toml
[endpoints]
base_api = "https://api.synthient.com/api/v4"
base_feeds = "https://feeds.synthient.com/v3"
base_grpc = "grpc.synthient.com:443"

[profiles.staging.endpoints]
base_api = "https://staging-api.example.com/api/v4"
base_feeds = "https://staging-feeds.example.com/v3"
base_grpc = "staging-grpc.example.com:443"
```

Use a profile:

```bash
synthient --profile staging status
```

Use a one-off config path:

```bash
synthient --config ./config.toml status
```

Global flags:

| Flag | Purpose |
| --- | --- |
| `--config <path>` | Read configuration from a non-default TOML file |
| `--profile <name>` | Apply endpoint overrides from `[profiles.<name>]` |
| `--no-color` | Disable ANSI styling |
| `--quiet`, `-q` | Suppress human progress output where supported |
| `--version`, `-v` | Print CLI version |

## Account and Permissions

Use `status` for a fast operational check:

```bash
synthient status
synthient status --format json
```

Use `account` when scripts need account data:

```bash
synthient account --format json | jq .
```

Use `scopes` to see which data surfaces are available to the active key:

```bash
synthient scopes
synthient scopes --format csv
```

Scope preflight checks are enabled for feed and stream commands that need feed-specific permissions. Use `--no-preflight` to skip the account lookup when you already know the key is scoped correctly or when minimizing an extra API call matters:

```bash
synthient feeds snapshots proxies --no-preflight
synthient stream proxies --duration 5s --no-preflight
```

Relevant scopes:

| Scope | Grants |
| --- | --- |
| `BASIC` | Account, IP lookup, batch IP lookup, and domain lookup |
| `PROXY_FEEDS` | Proxy snapshot list, metadata, checksum, schema, and download |
| `PROXY_FIREHOSE` | Proxy live stream |
| `ANONYMIZERS_FEED` | Anonymizer snapshot list, metadata, checksum, schema, and download |
| `ANONYMIZERS_STREAM` | Anonymizer live stream |
| `TORRENTS_FEED` | Torrent snapshot list, metadata, checksum, schema, and download |
| `TORRENTS_STREAM` | Torrent live stream |
| `HONEYPOT_HTTP_FEED` | HTTP honeypot snapshots |
| `HONEYPOT_HTTP_STREAM` | HTTP honeypot stream |
| `HONEYPOT_HTTPS_FEED` | HTTPS/TLS honeypot snapshots |
| `HONEYPOT_HTTPS_STREAM` | HTTPS/TLS honeypot stream |
| `HONEYPOT_DNS_FEED` | DNS honeypot snapshots |
| `HONEYPOT_DNS_STREAM` | DNS honeypot stream |
| `HONEYPOT_ADB_FEED` | ADB honeypot snapshots |
| `HONEYPOT_ADB_STREAM` | ADB honeypot stream |

## Lookup Workflows

Single IP:

```bash
synthient lookup 8.8.8.8
```

Batch IP lookup:

```bash
synthient lookup 8.8.8.8 1.1.1.1 9.9.9.9 --format json
```

Stdin lookup:

```bash
awk '{print $1}' access.log | sort -u | synthient lookup --format csv --output ip-intel.csv
```

Domain intelligence:

```bash
synthient lookup domain example.com
synthient lookup domain example.com --format json | jq .
```

The default `lookup` command and `lookup ip` are equivalent:

```bash
synthient lookup 8.8.8.8
synthient lookup ip 8.8.8.8
```

## Feed Snapshots

List supported streams:

```bash
synthient feeds streams
synthient feeds streams --format json
```

Current stream names:

| Stream | Aliases | Description |
| --- | --- | --- |
| `proxies` | `proxy` | Proxy IP observations across residential, datacenter, and mobile networks |
| `anonymizers` | `anonymizer` | VPN, Tor, private relay, and other anonymizer ranges |
| `torrents` | `torrent` | DHT and tracker peer sightings |
| `honeypot_http` | `helios_http`, `http` | HTTP request captures from Helios sensors |
| `honeypot_https` | `helios_https`, `https`, `tls` | TLS ClientHello captures from Helios sensors |
| `honeypot_dns` | `helios_dns`, `dns` | DNS resolution observations from Helios tunnels (snapshots only, no live stream) |
| `honeypot_adb` | `helios_adb`, `adb` | Android Debug Bridge shell commands captured by Helios sensors (snapshots only, no live stream) |

List snapshots:

```bash
synthient feeds snapshots proxies --limit 25
synthient feeds snapshots proxies --limit 25 --format json
```

Paginate:

```bash
PAGE=$(synthient feeds snapshots proxies --limit 25 --format json)
CURSOR=$(jq -r '.next_cursor // empty' <<<"$PAGE")
synthient feeds snapshots proxies --cursor "$CURSOR" --limit 25
```

Snapshot identifiers accepted by metadata, schema, checksum, and download commands:

```txt
latest
YYYY-MM-DD
YYYY-MM-DD/HH
```

For reproducible automation, prefer resolving a snapshot ID from `feeds snapshots` instead of hard-coding `latest`:

```bash
SNAPSHOT=$(synthient feeds snapshots proxies --limit 1 --format json | jq -r '.feeds[0].id')
synthient feeds meta proxies "$SNAPSHOT" --format json
```

Inspect metadata and schema:

```bash
synthient feeds meta proxies "$SNAPSHOT"
synthient feeds schema proxies "$SNAPSHOT"
synthient feeds checksum proxies "$SNAPSHOT"
```

Download and verify:

```bash
synthient feeds download proxies "$SNAPSHOT" proxies.parquet --verify --force
```

The download UI uses feed metadata size when available, so interactive terminals get a determinate progress bar with size, rate, elapsed time, and percentage.

The output file must end in `.parquet`. Without `--force`, existing files are not overwritten.

## Download Convenience Command

`synthient download` is a shorter wrapper around feed downloads.

Default wrapper behavior:

```bash
synthient download anonymizers-latest.parquet
```

That downloads the wrapper command's default stream with the default `latest` snapshot selection. For production jobs, prefer an explicit stream and snapshot.

Explicit stream and snapshot:

```bash
synthient download proxies "$SNAPSHOT" proxies.parquet --force
```

Date-based forms:

```bash
synthient download proxies proxies.parquet --date 2026-06-02
synthient download proxies proxies.parquet --date 2026-06-03 --hour 0
```

Flags:

| Flag | Purpose |
| --- | --- |
| `--date YYYY-MM-DD|latest`, `-d` | Snapshot date for the wrapper command |
| `--hour 0-23` | UTC hour for an hourly snapshot |
| `--force` | Overwrite an existing `.parquet` file |
| `--verify` | Verify the downloaded file against snapshot metadata checksum |
| `--silent`, `-s` | Suppress download output |
| `--no-preflight` | Skip feed scope preflight |

For scripts that already resolved a snapshot ID, prefer the explicit command:

```bash
synthient feeds download proxies "$SNAPSHOT" proxies.parquet --verify --force
```

## Live Streams

Stream commands emit newline-delimited JSON from live feed endpoints.

Basic stream:

```bash
synthient stream proxies
```

Capture a fixed number of matching events:

```bash
synthient stream proxies --filter type=RESIDENTIAL_PROXY --max-events 250 --output proxies.ndjson
```

Capture for a fixed duration:

```bash
synthient stream proxies --filter type=RESIDENTIAL_PROXY --duration 5s
```

Reconnect on server-side stream closure:

```bash
synthient stream proxies --reconnect --output proxies.ndjson
```

Pretty-print for terminal inspection:

```bash
synthient stream proxies --max-events 3 --pretty
```

Filter syntax:

```txt
--filter field=value
```

Filters are repeatable and all filters must match:

```bash
synthient stream proxies \
  --filter type=RESIDENTIAL_PROXY \
  --filter country_code=US \
  --duration 30s
```

Nested fields can be addressed with dot notation:

```bash
synthient stream proxies --filter asn=15169 --max-events 10
```

Because stream output is NDJSON, it is safe to pipe directly into `jq`, log processors, and queue producers:

```bash
synthient stream proxies --duration 5s \
  | jq -c 'select(.type == "RESIDENTIAL_PROXY")'
```

## gRPC Schema Introspection

Use gRPC reflection to inspect available protobuf descriptors:

```bash
synthient grpc schema
```

Request specific symbols:

```bash
synthient grpc schema synthient.v1.SynthientService
synthient grpc schema synthient.v1.SynthientService.StreamProxies
```

Write JSON:

```bash
synthient grpc schema synthient.v1.SynthientService --format json --output synthient-schema.json
```

Write a binary `FileDescriptorSet`:

```bash
synthient grpc schema synthient.v1.SynthientService --format binpb --output synthient.protoset
```

Use a custom endpoint:

```bash
synthient grpc schema --endpoint grpc.synthient.com:443
```

Use plaintext for local reflection servers:

```bash
synthient grpc schema --endpoint 127.0.0.1:50051 --plaintext
```

Include reflection service descriptors:

```bash
synthient grpc schema --include-reflection
```

Flags:

| Flag | Purpose |
| --- | --- |
| `--format text|json|binpb`, `-f` | Output representation |
| `--output -|file`, `-o` | Destination |
| `--endpoint <host:port>` | Override configured gRPC endpoint |
| `--timeout <duration>` | Reflection timeout, default `15s` |
| `--plaintext` | Connect without TLS |
| `--include-reflection` | Include gRPC reflection descriptors in output |

## MCP Server

`synthient mcp` runs a [Model Context Protocol](https://modelcontextprotocol.io) server that exposes Synthient lookups, feed metadata, live samples, and schemas as tools for MCP-compatible clients such as Claude Desktop. The server communicates over stdio.

Start the server:

```bash
synthient mcp
```

The server authenticates with the same credential lookup order as the rest of the CLI (`SYNTHIENT_API_KEY`, `.env`, then the OS keychain) and exits at startup if no key is found.

Exposed tools:

| Tool | Purpose |
| --- | --- |
| `lookup_ip` | Look up intelligence for one or more IP addresses |
| `lookup_domain` | Look up domain intelligence from Helios observations |
| `get_account` | Report organization, scopes, and lookup quota |
| `list_feed_streams` | List available feed streams with descriptions and aliases |
| `list_feed_snapshots` | List Parquet snapshots for a stream, with pagination |
| `feed_snapshot_meta` | Return snapshot metadata, checksum, size, row count, and schema |
| `sample_stream` | Collect a bounded sample of live events from a real-time feed |
| `grpc_schema` | Fetch protobuf descriptors through gRPC reflection |

Snapshot tools (`list_feed_snapshots`, `feed_snapshot_meta`) and `sample_stream` cover the streams that support those surfaces; the honeypot DNS and ADB feeds are snapshot-only and have no live sample.

Flags:

| Flag | Purpose |
| --- | --- |
| `--transport <stdio>` | Transport the server listens on; only `stdio` is supported |

### Claude Desktop

Add the server to `claude_desktop_config.json` (on macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "synthient": {
      "command": "synthient",
      "args": ["mcp"]
    }
  }
}
```

Use an absolute path to the binary if `synthient` is not on the launcher's `PATH` (for example `/Users/you/go/bin/synthient`). The keychain credential is used automatically. To authenticate from the environment instead, add an `env` block:

```json
{
  "mcpServers": {
    "synthient": {
      "command": "synthient",
      "args": ["mcp"],
      "env": { "SYNTHIENT_API_KEY": "..." }
    }
  }
}
```

Fully quit and reopen Claude Desktop after editing the config so it relaunches the server.

## Shell Completion

Generate completion scripts through Cobra:

```bash
synthient completion zsh > "${fpath[1]}/_synthient"
synthient completion bash > /usr/local/etc/bash_completion.d/synthient
synthient completion fish > ~/.config/fish/completions/synthient.fish
```

Open the generated shell-specific help:

```bash
synthient completion zsh --help
```

## Developer Recipes

Write account status to JSON and fail if the account is unreachable:

```bash
synthient status --format json \
  | jq -e '.account_reachable == true'
```

Export IP intelligence for a file of addresses:

```bash
sort -u ips.txt \
  | synthient lookup --format csv --output ip-intel.csv
```

Download the newest proxy snapshot and verify the checksum:

```bash
SNAPSHOT=$(synthient feeds snapshots proxies --limit 1 --format json | jq -r '.feeds[0].id')
synthient feeds download proxies "$SNAPSHOT" "proxies-$SNAPSHOT.parquet" --verify
```

Capture 250 residential proxy events quickly:

```bash
synthient stream proxies \
  --filter type=RESIDENTIAL_PROXY \
  --max-events 250 \
  --output residential-proxies.ndjson
```

Record a five-second stream sample and summarize with `jq`:

```bash
synthient stream proxies --filter type=RESIDENTIAL_PROXY --duration 5s \
  | jq -r '[.timestamp, .ip, .country_code, .provider] | @tsv'
```

Generate a protoset for downstream tooling:

```bash
synthient grpc schema synthient.v1.SynthientService \
  --format binpb \
  --output synthient.protoset
```

## Local Development

Run the CLI from source:

```bash
go run ./cmd status
```

Build:

```bash
go build -o /tmp/synthient-cli ./cmd
```

Test and vet:

```bash
go test ./...
go vet ./...
```

Full local audit:

```bash
go test ./...
go vet ./...
go build -o /tmp/synthient-cli ./cmd
```

Smoke-check production commands with a live key:

```bash
SYNTHIENT_API_KEY="..." /tmp/synthient-cli status --format json
SYNTHIENT_API_KEY="..." /tmp/synthient-cli lookup 8.8.8.8 --format json
SYNTHIENT_API_KEY="..." /tmp/synthient-cli feeds snapshots proxies --limit 3 --no-preflight
SNAPSHOT=$(SYNTHIENT_API_KEY="..." /tmp/synthient-cli feeds snapshots proxies --limit 1 --format json --no-preflight | jq -r '.feeds[0].id')
SYNTHIENT_API_KEY="..." /tmp/synthient-cli feeds schema proxies "$SNAPSHOT" --no-preflight
SYNTHIENT_API_KEY="..." /tmp/synthient-cli stream proxies --filter type=RESIDENTIAL_PROXY --duration 5s --no-preflight
SYNTHIENT_API_KEY="..." /tmp/synthient-cli grpc schema synthient.v1.SynthientService
```

## Troubleshooting

No credentials found:

```txt
Set SYNTHIENT_API_KEY, add SYNTHIENT_API_KEY to .env, or run synthient auth.
```

Colors are missing:

```txt
Check NO_COLOR, --no-color, terminal type, and whether output is redirected.
```

Download refuses to overwrite:

```txt
Pass --force or choose a new .parquet filename.
```

Download progress is hidden:

```txt
Progress is shown only for interactive terminals. Redirected output and --quiet use non-interactive behavior.
```

Missing feed or stream access:

```txt
Run synthient scopes and compare the required feed or stream scope. Use --no-preflight only when you intentionally want to skip the extra account check.
```

Snapshot lookup fails:

```txt
List snapshots first and use the returned id. For automation, prefer:
SNAPSHOT=$(synthient feeds snapshots <stream> --limit 1 --format json | jq -r '.feeds[0].id')
```

gRPC schema connection fails:

```txt
Check the endpoint, TLS mode, and timeout. Use --plaintext only for local or explicitly plaintext reflection servers.
```

MCP server tools do not appear in the client:

```txt
Confirm the command path resolves, fully restart the client so it relaunches the server, and check that the binary is current (synthient --version). Older binaries may not expose every tool.
```

# synthient/cli

Synthient's official CLI tool. Currently in beta.

## Install

| System  | Install Command                                                                            |
| ------- | ------------------------------------------------------------------------------------------ |
| macOS   | `brew install synthient/tap/synthient` or `go install github.com/synthient/cli/cmd@latest` |
| Linux   | `go install github.com/synthient/cli/cmd@latest`                                           |
| Windows | `go install github.com/synthient/cli/cmd@latest`                                           |

## Basic Commands

### `synthient auth`

![auth demo video](./demos/auth.gif)

Authenticate the CLI with your Synthient API key. This API key will then be stored in your system's keychain for secure storage. If you want to authenticate a different way there are two other options:

1. Provide your API key as an environment variable with the key being `SYNTHIENT_API_KEY`.
2. Store it in a `.env` file in the current directory with the key `SYNTHIENT_API_KEY`.

### `synthient lookup`

![lookup demo video](./demos/lookup.gif)

> [!NOTE]
> You need to be [authenticated](#synthient-auth) to run this command.

Lookup a given IP address. Works with piped input and/or multiple IP address as seen here:

```bash
synthient lookup 213.149.183.127
```

```bash
synthient lookup 213.149.183.127 168.205.174.84
```

```bash
echo "213.149.183.127 168.205.174.84" | synthient lookup
```

Here is a full look into all of the options for the `lookup` command:

```txt
Usage:
  synthient lookup [flags]

Flags:
  -f, --format string   Output format [text|json|csv] (default "text")
  -h, --help            help for lookup
  -o, --output string   Where to write output: '-' for stdout, or a file path (e.g. 'lookup.json' or 'lookup.csv) (default "-")
```

## Feeds

### Streaming (`synthient stream`)

![stream demo video](./demos/stream.gif)

> [!NOTE]
> You need to be [authenticated](#synthient-auth) to run this command.

Stream feed events as newline-delimited JSON to stdout. The first argument selects the feed:

```bash
synthient stream anonymizer
```

```bash
synthient stream torrent
```

```bash
synthient stream proxy
```

### Downloading (`synthient download`)

![download demo video](./demos/download.gif)

> [!NOTE]
> You need to be [authenticated](#synthient-auth) to run this command.

Download an anonymizer feed snapshot as a Parquet file. The filename argument must end in `.parquet`:

```bash
synthient download anonymizers.parquet
```

To download a specific date's snapshot instead of the latest:

```bash
synthient download anonymizers.parquet --date 2026-05-14
```

Here is a full look into all of the options for the `download` command:

```txt
Usage:
  synthient download [flags]

Flags:
  -d, --date string   Snapshot date to download (YYYY-MM-DD or 'latest') (default "latest")
  -h, --help          help for download
  -s, --silent        Do not output when downloading
```

## Configuration

Configuration for the synthient CLI is located in `~/.config/synthient/config.toml`. Right now the customizable fields are the hosts for API and feeds data:

```toml
[endpoints]
base_api = "https://customhost.com"
base_feeds = "https://otherhost.com"
```

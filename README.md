# synthient/cli

Synthient's official CLI tool. Currently in beta.

## Install

| System  | Install Command                                                                 |
| ------- | ------------------------------------------------------------------------------- |
| macOS   | `brew install synthient/tap/synthient` or `go install github.com/synthient/cli` |
| Linux   | `go install github.com/synthient/cli`                                           |
| Windows | `go install github.com/synthient/cli`                                           |

## Basic Commands

### `synthient auth`

![auth demo tape](./demos/auth.gif)

Authenticate the CLI with your Synthient API key. This API key will then be stored in your system's keychain for secure storage. If you want to authenticate a different way there are two other options:

1. Provide your API key as an environment variable with the key being `SYNTHIENT_API_KEY`.
2. Store it in a `.env` file in the current directory with the key `SYNTHIENT_API_KEY`.

### `synthient lookup`

<details>
  <summary>DEMO VIDEO</summary>
  <img src="./demos/lookup.gif">
</details>

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

<details>
  <summary>DEMO VIDEO</summary>
  <img src="./demos/stream.gif">
</details>

> [!NOTE]
> You need to be [authenticated](#synthient-auth) to run this command.

Want to stream data from a anonymizer feed? Do this using the synthient CLI with the following command:

```bash
synthient stream \
   --provider IPIDEA \
   --type RESIDENTIAL_PROXY \
   --last_observed 24hr \
   --format CSV \
   --country_code US \
   --order desc
```

This will then stream the data and output it to the standard out as it comes in.

### Downloading (`synthient download`)

<details>
  <summary>DEMO VIDEO</summary>
  <img src="./demos/download.gif">
</details>

> [!NOTE]
> You need to be [authenticated](#synthient-auth) to run this command.

The same can be done for [streaming](#streaming) but instead it goes directly into a file. This can be done by using the `download` command with the same flags available but with the file specified as an argument (e.g. `feed.csv`):

```bash
synthient download feed.csv \
   --provider BIRDPROXIES \
   --type RESIDENTIAL_PROXY \
   --last_observed 7D \
   --format CSV \
   --country_code US \
   --order desc
```

## Configuration

Configuration for the synthient CLI is located in `~/.config/synthient/config.toml`. Right now the customizable fields are the hosts for API and feeds data:

```toml
[endpoints]
base_api = "https://customhost.com"
base_feeds = "https://otherhost.com"
```

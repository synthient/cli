#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

: "${SYNTHIENT_API_KEY:?SYNTHIENT_API_KEY is required}"

for command in go vhs ffmpeg jq awk; do
	if ! command -v "$command" >/dev/null 2>&1; then
		echo "$command is required" >&2
		exit 1
	fi
done

mkdir -p dist
go build -o dist/synthient ./cmd

export PATH="$PWD/dist:$PATH"
unset NO_COLOR
export TERM=xterm-256color
export COLORTERM=truecolor

rm -f demos/proxies-latest.parquet demos/proxies-latest.parquet.part

for tape in auth lookup feeds stream download grpc; do
	vhs "demos/$tape.tape"
	ffmpeg -y -loglevel error -i "demos/$tape.gif" -movflags faststart -pix_fmt yuv420p "demos/$tape.mp4"
done

rm -f demos/proxies-latest.parquet demos/proxies-latest.parquet.part

for asset in \
	demos/auth.gif demos/auth.mp4 \
	demos/lookup.gif demos/lookup.mp4 \
	demos/feeds.gif demos/feeds.mp4 \
	demos/stream.gif demos/stream.mp4 \
	demos/download.gif demos/download.mp4 \
	demos/grpc.gif demos/grpc.mp4; do
	if [ ! -s "$asset" ]; then
		echo "$asset was not generated" >&2
		exit 1
	fi
done

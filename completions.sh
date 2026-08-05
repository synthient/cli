#!/bin/sh
set -e
rm -rf completions && mkdir completions
for sh in bash zsh fish; do
	go run ./cmd/synthient completion "$sh" > "completions/synthient.$sh"
done

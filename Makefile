completion:
	go run cmd/synthient.go completion bash > ../homebrew-tap/completions/synthient.bash
	go run cmd/synthient.go completion fish > ../homebrew-tap/completions/synthient.fish
	go run cmd/synthient.go completion zsh > ../homebrew-tap/completions/synthient.zsh
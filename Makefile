completion:
	go run cmd/synthient.go completion bash > completions/synthient.bash
	go run cmd/synthient.go completion fish > completions/synthient.fish
	go run cmd/synthient.go completion zsh > completions/synthient.zsh
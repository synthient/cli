package grpcschema

import "testing"

func TestNormalizeEndpoint(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		endpoint   string
		serverName string
	}{
		{
			name:       "default",
			input:      "",
			endpoint:   "grpc.synthient.com:443",
			serverName: "grpc.synthient.com",
		},
		{
			name:       "host without port",
			input:      "grpc.example.com",
			endpoint:   "grpc.example.com:443",
			serverName: "grpc.example.com",
		},
		{
			name:       "host with port",
			input:      "grpc.example.com:8443",
			endpoint:   "grpc.example.com:8443",
			serverName: "grpc.example.com",
		},
		{
			name:       "url",
			input:      "https://grpc.example.com:443",
			endpoint:   "grpc.example.com:443",
			serverName: "grpc.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint, serverName, err := NormalizeEndpoint(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			if endpoint != tt.endpoint {
				t.Fatalf("endpoint = %q, want %q", endpoint, tt.endpoint)
			}
			if serverName != tt.serverName {
				t.Fatalf("serverName = %q, want %q", serverName, tt.serverName)
			}
		})
	}
}

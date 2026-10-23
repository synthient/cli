package conf

import "github.com/synthient/go-synthient/v2"

func (config Config) ApplyToClient(client *synthient.Client) {
	if config.BaseApiURL != nil {
		client.BaseAPI = *config.BaseApiURL
	}
}

func (config Config) GRPCEndpoint(defaultEndpoint string) string {
	if config.Endpoints.BaseGRPC != "" {
		return config.Endpoints.BaseGRPC
	}
	return defaultEndpoint
}

package conf

import "github.com/synthient/synthient-go"

func (config Config) ClientOptions() []synthient.Option {
	if config.BaseApiURL == nil {
		return nil
	}
	return []synthient.Option{synthient.WithBaseURL(config.BaseApiURL.String())}
}

func (config Config) GRPCEndpoint(defaultEndpoint string) string {
	if config.Endpoints.BaseGRPC != "" {
		return config.Endpoints.BaseGRPC
	}
	return defaultEndpoint
}

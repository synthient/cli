package conf

import "github.com/synthient/go-synthient/v2"

func (config Config) ApplyToClient(client *synthient.Client) {
	if config.BaseApiURL != nil {
		client.BaseAPI = *config.BaseApiURL
	}
	if config.BaseFeedsURL != nil {
		client.BaseFeeds = *config.BaseFeedsURL
	}
}

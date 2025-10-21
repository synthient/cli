package synthient

import (
	"fmt"
	"net/http"

	"github.com/synthient/cli/internal/output"
)

func (client *Client) LookupIP(ip string) (LookupResponse, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("https://v3api.synthient.com/api/v3/lookup/ip/%s", ip),
		nil,
	)
	if err != nil {
		return LookupResponse{}, fmt.Errorf("%w creating request failed", err)
	}

	resp, err := synthientRequest[LookupResponse](client, req)
	if err != nil {
		return LookupResponse{}, fmt.Errorf("%w sending lookup request failed", err)
	}
	return resp, nil
}

func (response LookupResponse) Output() {
	blocks := []output.Block{
		{
			Name: "Location",
			Values: map[string]any{
				"City":     response.Location.City,
				"Timezone": response.Location.Timezone,
			},
		},
	}

	for i, block := range blocks {
		block.Output()
		if i+1 != len(blocks) {
			fmt.Println()
		}
	}
}

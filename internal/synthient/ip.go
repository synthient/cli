package synthient

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/charmbracelet/lipgloss"
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
	fmt.Println(output.SYNTHIENT_COLOR.Bold(true).Render("IP")+":", response.IP)
	fmt.Println()

	blocks := []output.Block{
		{
			Name: "Network",
			Values: []output.BlockValue{
				{Key: "Asn", Value: response.Network.Asn},
				{Key: "ISP", Value: response.Network.Isp},
				{Key: "Type", Value: response.Network.Type},
			},
		},
		{
			Name: "Location",
			Values: []output.BlockValue{
				{Key: "City", Value: response.Location.City},
				{Key: "State", Value: response.Location.State},
				{Key: "Country", Value: response.Location.Country},
				{Key: "Timezone", Value: response.Location.Timezone},
				{Key: "Longitude", Value: response.Location.Longitude},
				{Key: "Latitude", Value: response.Location.Latitude},
				{Key: "Geo Hash", Value: response.Location.GeoHash},
			},
		},
		{
			Name: "IP Data",
			Values: []output.BlockValue{
				{Key: "Device Count", Value: response.IPData.DeviceCount},
				{Key: "Categories", Value: strings.Join(response.IPData.Categories, ", ")},
				{Key: "IP Risk", Value: response.IPData.IPRisk},
			},
		},
	}

	for i, block := range blocks {
		block.Output(0)
		if i+1 != len(blocks) {
			fmt.Println()
		}
	}

	if len(response.IPData.Enriched) != 0 {
		headerStyle := lipgloss.NewStyle().Bold(true)
		fmt.Println()
		fmt.Println(headerStyle.Render("Proxy Providers"))
		for i, enrichedProxyData := range response.IPData.Enriched {
			fmt.Printf("   %d. %s\n", i+1, enrichedProxyData.Provider)
			block := output.Block{
				Values: []output.BlockValue{
					{Key: "Type", Value: enrichedProxyData.Type},
					{Key: "Last Seen", Value: enrichedProxyData.LastSeen},
				},
			}
			block.Output(5)
		}
	}
}

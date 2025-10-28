package synthient

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/synthient/cli/internal/output"
	"go.mattglei.ch/timber"
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
		if errors.Is(err, ErrApi) {
			timber.ErrorMsg(
				fmt.Sprintf(
					"Failed to look up IP '%s': %s",
					ip,
					strings.TrimSpace(strings.TrimPrefix(err.Error(), ErrApi.Error())),
				),
			)
			os.Exit(1)
		}
		return LookupResponse{}, fmt.Errorf("%w sending lookup request failed", err)
	}

	return resp, nil
}

func (r LookupResponse) Output(out *os.File, styles output.Styles, spacing bool) {
	output.WriteLine(out, styles.SynthientColor.Bold(true).Render("IP")+":", r.IP)
	if spacing {
		output.WriteLine(out)
	}

	blocks := []output.Block{
		{
			Name: "Network",
			Values: []output.BlockValue{
				{Key: "Asn", Value: r.Network.Asn},
				{Key: "ISP", Value: r.Network.Isp},
				{Key: "Type", Value: r.Network.Type},
			},
		},
		{
			Name: "Location",
			Values: []output.BlockValue{
				{Key: "City", Value: r.Location.City},
				{Key: "State", Value: r.Location.State},
				{Key: "Country", Value: r.Location.Country},
				{Key: "Timezone", Value: r.Location.Timezone},
				{Key: "Longitude", Value: r.Location.Longitude},
				{Key: "Latitude", Value: r.Location.Latitude},
				{Key: "Geo Hash", Value: r.Location.GeoHash},
			},
		},
		{
			Name: "IP Data",
			Values: []output.BlockValue{
				{Key: "Device Count", Value: r.IPData.DeviceCount},
				{Key: "Categories", Value: strings.Join(r.IPData.Categories, ", ")},
				{Key: "IP Risk", Value: r.IPData.IPRisk},
			},
		},
	}

	for i, block := range blocks {
		block.Output(out, styles, 0)
		if spacing && i+1 != len(blocks) {
			output.WriteLine(out)
		}
	}

	if len(r.IPData.Enriched) != 0 {
		headerStyle := lipgloss.NewStyle().Bold(true).Renderer(lipgloss.NewRenderer(out))
		if spacing {
			output.WriteLine(out)
		}
		output.WriteLine(out, headerStyle.Render("Proxy Providers"))
		for i, enrichedProxyData := range r.IPData.Enriched {
			output.WriteLine(out, fmt.Sprintf("   %d. %s", i+1, enrichedProxyData.Provider))
			block := output.Block{
				Values: []output.BlockValue{
					{Key: "Type", Value: enrichedProxyData.Type},
					{Key: "Last Seen", Value: enrichedProxyData.LastSeen},
				},
			}
			block.Output(out, styles, 5)
		}
	}
}

func (r LookupResponse) OutputCSV(writer *csv.Writer) {
	enrichedData, err := json.Marshal(r.IPData.Enriched)
	if err != nil {
		timber.Fatal(err, "failed to marshal JSON data for enriched data")
	}

	err = writer.Write([]string{
		r.IP,

		strconv.Itoa(r.Network.Asn),
		r.Network.Isp,
		r.Network.Type,

		r.Location.City,
		r.Location.State,
		r.Location.Country,
		r.Location.Timezone,
		strconv.FormatFloat(r.Location.Longitude, 'f', -1, 64),
		strconv.FormatFloat(r.Location.Latitude, 'f', -1, 64),
		r.Location.GeoHash,

		strings.Join(r.IPData.Behavior, "|"),
		strings.Join(r.IPData.Categories, "|"),
		strconv.Itoa(r.IPData.IPRisk),
		string(enrichedData),
	})
	if err != nil {
		timber.Fatal(err, "failed to write csv row for ip")
	}
	writer.Flush()
}

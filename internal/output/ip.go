package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/synthient/go-synthient"
	"go.mattglei.ch/timber"
)

func IP(ip synthient.IP, out *os.File, styles Styles, spacing bool) {
	WriteLine(out, styles.SynthientColor.Bold(true).Render("IP")+":", ip.IP)
	if spacing {
		WriteLine(out)
	}

	var categories string
	if len(ip.IPData.Categories) == 0 {
		categories = "None"
	} else {
		categories = strings.Join(ip.IPData.Categories, ", ")
	}

	blocks := []Block{
		{
			Name: "Network",
			Values: []BlockValue{
				{Key: "Asn", Value: ip.Network.Asn},
				{Key: "ISP", Value: ip.Network.Isp},
				{Key: "Type", Value: ip.Network.Type},
			},
		},
		{
			Name: "Location",
			Values: []BlockValue{
				{Key: "City", Value: ip.Location.City},
				{Key: "State", Value: ip.Location.State},
				{Key: "Country", Value: ip.Location.Country},
				{Key: "Timezone", Value: ip.Location.Timezone},
				{Key: "Longitude", Value: ip.Location.Longitude},
				{Key: "Latitude", Value: ip.Location.Latitude},
				{Key: "Geo Hash", Value: ip.Location.GeoHash},
			},
		},
		{
			Name: "IP Data",
			Values: []BlockValue{
				{Key: "Device Count", Value: ip.IPData.DeviceCount},
				{Key: "Categories", Value: categories},
				{Key: "IP Risk", Value: ip.IPData.IPRisk},
			},
		},
	}

	for i, block := range blocks {
		block.Output(out, styles, 0)
		if spacing && i+1 != len(blocks) {
			WriteLine(out)
		}
	}

	if len(ip.IPData.Enriched) != 0 {
		headerStyle := lipgloss.NewStyle().Bold(true).Renderer(lipgloss.NewRenderer(out))
		if spacing {
			WriteLine(out)
		}
		WriteLine(out, headerStyle.Render("Proxy Providers"))
		for i, enrichedProxyData := range ip.IPData.Enriched {
			WriteLine(out, fmt.Sprintf("   %d. %s", i+1, enrichedProxyData.Provider))
			block := Block{
				Values: []BlockValue{
					{Key: "Type", Value: enrichedProxyData.Type},
					{Key: "Last Seen", Value: enrichedProxyData.LastSeen},
				},
			}
			block.Output(out, styles, 5)
		}
	}
}

func IpCSV(ip synthient.IP, writer *csv.Writer) {
	enrichedData, err := json.Marshal(ip.IPData.Enriched)
	if err != nil {
		timber.Fatal(err, "failed to marshal JSON data for enriched data")
	}

	err = writer.Write([]string{
		ip.IP,

		strconv.Itoa(ip.Network.Asn),
		ip.Network.Isp,
		ip.Network.Type,

		ip.Location.City,
		ip.Location.State,
		ip.Location.Country,
		ip.Location.Timezone,
		strconv.FormatFloat(ip.Location.Longitude, 'f', -1, 64),
		strconv.FormatFloat(ip.Location.Latitude, 'f', -1, 64),
		ip.Location.GeoHash,

		strings.Join(ip.IPData.Behavior, "|"),
		strings.Join(ip.IPData.Categories, "|"),
		strconv.Itoa(ip.IPData.IPRisk),
		string(enrichedData),
	})
	if err != nil {
		timber.Fatal(err, "failed to write csv row for ip")
	}
	writer.Flush()
}

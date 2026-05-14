package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/synthient/go-synthient/v2"
	"go.mattglei.ch/timber"
)

func IP(ip synthient.IP, out *os.File, styles Styles, spacing bool) {
	WriteLine(out, styles.SynthientColor.Bold(true).Render("IP")+":", ip.IP)
	if spacing {
		WriteLine(out)
	}

	var categories string
	if len(ip.Intelligence.Categories) == 0 {
		categories = "None"
	} else {
		categories = strings.Join(ip.Intelligence.Categories, ", ")
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
			Name: "Intelligence",
			Values: []BlockValue{
				{Key: "Categories", Value: categories},
				{Key: "Risk Score", Value: ip.Intelligence.RiskScore},
			},
		},
	}

	for i, block := range blocks {
		block.Output(out, styles, 0)
		if spacing && i+1 != len(blocks) {
			WriteLine(out)
		}
	}

	if len(ip.Intelligence.Providers) != 0 {
		headerStyle := lipgloss.NewStyle().Bold(true).Renderer(lipgloss.NewRenderer(out))
		if spacing {
			WriteLine(out)
		}
		WriteLine(out, headerStyle.Render("Proxy Providers"))
		for i, provider := range ip.Intelligence.Providers {
			WriteLine(out, fmt.Sprintf("   %d. %s", i+1, provider.Provider))
			block := Block{
				Values: []BlockValue{
					{Key: "Type", Value: provider.Type},
					{Key: "Last Seen", Value: time.Unix(provider.LastSeen, 0).UTC().Format(time.RFC3339)},
				},
			}
			block.Output(out, styles, 5)
		}
	}
}

func IpCSV(ip synthient.IP, writer *csv.Writer) {
	providersData, err := json.Marshal(ip.Intelligence.Providers)
	if err != nil {
		timber.Fatal(err, "failed to marshal JSON data for providers data")
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

		strings.Join(ip.Intelligence.Behavior, "|"),
		strings.Join(ip.Intelligence.Categories, "|"),
		strconv.Itoa(ip.Intelligence.RiskScore),
		string(providersData),
	})
	if err != nil {
		timber.Fatal(err, "failed to write csv row for ip")
	}
	writer.Flush()
}

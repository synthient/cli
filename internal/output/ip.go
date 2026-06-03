package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/synthient/go-synthient/v2"
	"go.mattglei.ch/timber"
)

func IP(ip synthient.IP, out *os.File, styles Styles, spacing bool) {
	Header(out, styles, "IP", ip.IP)
	Divider(out, styles)

	blocks := []Block{
		{
			Name: "Network",
			Values: []BlockValue{
				{Key: "ASN", Value: ip.Network.Asn},
				{Key: "ISP", Value: ip.Network.Isp},
				{Key: "Type", Value: ip.Network.Type},
				{Key: "Org", Value: ip.Network.Org},
				{Key: "Domain", Value: ip.Network.Domain},
				{Key: "Abuse Email", Value: ip.Network.AbuseEmail},
				{Key: "Abuse Phone", Value: ip.Network.AbusePhone},
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
				{Key: "Risk Score", Value: ip.Intelligence.RiskScore},
				{Key: "Behavior", Value: ip.Intelligence.Behavior},
				{Key: "Categories", Value: ip.Intelligence.Categories},
			},
		},
	}

	if len(ip.Intelligence.Devices) != 0 {
		devices := Block{Name: "Devices"}
		for _, device := range ip.Intelligence.Devices {
			label := device.OS
			if device.Version != "" {
				label = fmt.Sprintf("%s %s", label, device.Version)
			}
			devices.Values = append(devices.Values, BlockValue{Key: label, Value: "Observed"})
		}
		blocks = append(blocks, devices)
	}

	if len(ip.Intelligence.Providers) != 0 {
		providers := Block{Name: "Proxy Providers"}
		for _, provider := range ip.Intelligence.Providers {
			value := provider.Type
			if provider.LastSeen > 0 {
				value = fmt.Sprintf("%s, last seen %s", value, time.Unix(provider.LastSeen, 0).UTC().Format(time.RFC3339))
			}
			providers.Values = append(providers.Values, BlockValue{Key: provider.Provider, Value: value})
		}
		blocks = append(blocks, providers)
	}

	for i, block := range blocks {
		block.Output(out, styles, 0, i+1 == len(blocks))
		if spacing && i+1 != len(blocks) {
			Divider(out, styles)
		}
	}
}

func IpCSV(ip synthient.IP, writer *csv.Writer) {
	providersData, err := json.Marshal(ip.Intelligence.Providers)
	if err != nil {
		timber.Fatal(err, "failed to marshal JSON data for providers data")
	}
	devicesData, err := json.Marshal(ip.Intelligence.Devices)
	if err != nil {
		timber.Fatal(err, "failed to marshal JSON data for devices data")
	}

	err = writer.Write([]string{
		ip.IP,

		strconv.Itoa(ip.Network.Asn),
		ip.Network.Isp,
		ip.Network.Type,
		ip.Network.Org,
		ip.Network.Domain,
		ip.Network.AbuseEmail,
		ip.Network.AbusePhone,

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
		string(devicesData),
		string(providersData),
	})
	if err != nil {
		timber.Fatal(err, "failed to write csv row for ip")
	}
	writer.Flush()
}

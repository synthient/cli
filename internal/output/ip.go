package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/synthient/synthient-go"
	"go.mattglei.ch/timber"
)

func IP(ip synthient.IPInfo, out *os.File, styles Styles, spacing bool) {
	network, location, intelligence := orZero(ip.Network), orZero(ip.Location), orZero(ip.Intelligence)
	Header(out, styles, "IP", ip.IP.String())
	Divider(out, styles)

	blocks := []Block{
		{
			Name: "Network",
			Values: []BlockValue{
				{Key: "ASN", Value: network.ASN},
				{Key: "ISP", Value: network.ISP},
				{Key: "Type", Value: string(network.Type)},
				{Key: "Org", Value: network.Org},
				{Key: "Domain", Value: network.Domain},
				{Key: "Abuse Email", Value: network.AbuseEmail},
				{Key: "Abuse Phone", Value: network.AbusePhone},
			},
		},
		{
			Name: "Location",
			Values: []BlockValue{
				{Key: "City", Value: location.City},
				{Key: "State", Value: location.State},
				{Key: "Country", Value: location.Country},
				{Key: "Timezone", Value: location.Timezone},
				{Key: "Longitude", Value: location.Longitude},
				{Key: "Latitude", Value: location.Latitude},
				{Key: "Geo Hash", Value: location.GeoHash},
			},
		},
		{
			Name: "Intelligence",
			Values: []BlockValue{
				{Key: "Risk Score", Value: intelligence.RiskScore},
				{Key: "Behavior", Value: intelligence.Behavior},
				{Key: "Categories", Value: intelligence.Categories},
			},
		},
	}

	if len(intelligence.Devices) != 0 {
		devices := Block{Name: "Devices"}
		for _, device := range intelligence.Devices {
			label := string(device.OS)
			if device.Version != "" {
				label = fmt.Sprintf("%s %s", label, device.Version)
			}
			devices.Values = append(devices.Values, BlockValue{Key: label, Value: "Observed"})
		}
		blocks = append(blocks, devices)
	}

	if len(intelligence.Providers) != 0 {
		providers := Block{Name: "Proxy Providers"}
		for _, provider := range intelligence.Providers {
			value := provider.Type
			if provider.LastSeen > 0 {
				value = fmt.Sprintf("%s, last seen %s", value, provider.LastSeen.Time().Format(time.RFC3339))
			}
			providers.Values = append(providers.Values, BlockValue{Key: provider.Name, Value: value})
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

func IpCSV(ip synthient.IPInfo, writer *csv.Writer) {
	network, location, intelligence := orZero(ip.Network), orZero(ip.Location), orZero(ip.Intelligence)
	providersData, err := json.Marshal(intelligence.Providers)
	if err != nil {
		timber.Fatal(err, "failed to marshal JSON data for providers data")
	}
	devicesData, err := json.Marshal(intelligence.Devices)
	if err != nil {
		timber.Fatal(err, "failed to marshal JSON data for devices data")
	}

	err = writer.Write([]string{
		ip.IP.String(),

		strconv.FormatInt(network.ASN, 10),
		network.ISP,
		string(network.Type),
		network.Org,
		network.Domain,
		network.AbuseEmail,
		network.AbusePhone,

		location.City,
		location.State,
		location.Country,
		location.Timezone,
		strconv.FormatFloat(location.Longitude, 'f', -1, 64),
		strconv.FormatFloat(location.Latitude, 'f', -1, 64),
		location.GeoHash,

		strings.Join(intelligence.Behavior, "|"),
		strings.Join(intelligence.Categories, "|"),
		strconv.Itoa(intelligence.RiskScore),
		string(devicesData),
		string(providersData),
	})
	if err != nil {
		timber.Fatal(err, "failed to write csv row for ip")
	}
	writer.Flush()
}

func orZero[T any](value *T) T {
	if value == nil {
		var zero T
		return zero
	}
	return *value
}

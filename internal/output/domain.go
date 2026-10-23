package output

import (
	"fmt"
	"os"
	"time"

	"github.com/synthient/synthient-go"
)

func Domain(domain synthient.DomainInfo, out *os.File, styles Styles) {
	Header(out, styles, "Domain", domain.Domain)
	Divider(out, styles)

	blocks := []Block{
		{
			Name: "Status",
			Values: []BlockValue{
				{Key: "State", Value: string(domain.Status)},
				{Key: "Events 24h", Value: domain.Stats.Events24h},
				{Key: "Events 30d", Value: domain.Stats.TotalEvents30d},
				{Key: "Unique IPs 24h", Value: domain.UniqueIPs.Value24h},
				{Key: "Unique IPs 30d", Value: domain.UniqueIPs.Value30d},
			},
		},
		{
			Name: "Activity",
			Values: []BlockValue{
				{Key: "Peak Hour", Value: domain.ActivityStats.PeakHour},
				{Key: "Quiet Hour", Value: domain.ActivityStats.QuietHour},
				{Key: "Median Per Hour", Value: domain.ActivityStats.MedianPerHour},
				{Key: "P95 Per Hour", Value: domain.ActivityStats.P95PerHour},
				{Key: "Cadence", Value: string(domain.ActivityStats.Cadence)},
			},
		},
	}

	if domain.TopASN.ASN != 0 || domain.TopASN.Events != 0 {
		blocks = append(blocks, Block{
			Name: "Top ASN",
			Values: []BlockValue{
				{Key: "ASN", Value: domain.TopASN.ASN},
				{Key: "Events", Value: domain.TopASN.Events},
			},
		})
	}

	if len(domain.TopSubdomains) != 0 {
		block := Block{Name: "Top Subdomains"}
		for _, item := range domain.TopSubdomains {
			block.Values = append(block.Values, BlockValue{Key: item.Subdomain, Value: item.Count})
		}
		blocks = append(blocks, block)
	}

	if len(domain.TopPorts) != 0 {
		block := Block{Name: "Top Ports"}
		for _, item := range domain.TopPorts {
			block.Values = append(block.Values, BlockValue{Key: fmt.Sprint(item.Port), Value: item.Count})
		}
		blocks = append(blocks, block)
	}

	if len(domain.GeoDistribution) != 0 {
		block := Block{Name: "Geo Distribution"}
		for _, item := range domain.GeoDistribution {
			block.Values = append(block.Values, BlockValue{Key: item.CountryCode, Value: fmt.Sprintf("%d events, %d unique IPs", item.Events, item.UniqueIPs)})
		}
		blocks = append(blocks, block)
	}

	if len(domain.RecentEvents) != 0 {
		block := Block{Name: "Recent Events"}
		for _, item := range domain.RecentEvents {
			label := item.Timestamp.Time().Format(time.RFC3339)
			block.Values = append(block.Values, BlockValue{Key: label, Value: fmt.Sprintf("%s %s:%d %s", item.SourceIP, item.TargetSubdomain, item.Port, item.CountryCode)})
		}
		blocks = append(blocks, block)
	}

	for i, block := range blocks {
		block.Output(out, styles, 0, i+1 == len(blocks))
		if i+1 != len(blocks) {
			Divider(out, styles)
		}
	}
}

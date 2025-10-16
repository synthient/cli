package synthient

import (
	"fmt"
	"net/http"
)

func LookupIP(client *Client, ip string) (LookupResponse, error) {
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

type LookupResponse struct {
	IP      string `json:"ip"`
	Network struct {
		Asn        int         `json:"asn"`
		Isp        string      `json:"isp"`
		Type       string      `json:"type"`
		Org        interface{} `json:"org"`
		AbuseEmail interface{} `json:"abuse_email"`
		AbusePhone interface{} `json:"abuse_phone"`
		Domain     interface{} `json:"domain"`
	} `json:"network"`
	Location struct {
		Country   string  `json:"country"`
		State     string  `json:"state"`
		City      string  `json:"city"`
		Timezone  string  `json:"timezone"`
		Longitude float64 `json:"longitude"`
		Latitude  float64 `json:"latitude"`
		GeoHash   string  `json:"geo_hash"`
	} `json:"location"`
	IPData struct {
		DeviceCount int           `json:"device_count"`
		Devices     []interface{} `json:"devices"`
		Behavior    []string      `json:"behavior"`
		Categories  []string      `json:"categories"`
		Enriched    []struct {
			Provider string `json:"provider"`
			Type     string `json:"type"`
			LastSeen string `json:"last_seen"`
		} `json:"enriched"`
		IPRisk int `json:"ip_risk"`
	} `json:"ip_data"`
}

package synthient

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
		City      string  `json:"city"`
		State     string  `json:"state"`
		Country   string  `json:"country"`
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
		IPRisk      int           `json:"ip_risk"`
		Enriched    []struct {
			Provider string `json:"provider"`
			Type     string `json:"type"`
			LastSeen string `json:"last_seen"`
		} `json:"enriched"`
	} `json:"ip_data"`
}

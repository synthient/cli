package process

import (
	"encoding/csv"

	"github.com/charmbracelet/huh"
	"github.com/synthient/cli/internal/output"
	"go.mattglei.ch/timber"
)

func GetIpColumn(cr *csv.Reader) int {
	row, err := cr.Read()
	if err != nil {
		timber.Fatal(err, "failed to read header of csv file")
	}
	options := []huh.Option[int]{}
	for i, column := range row {
		options = append(options, huh.NewOption(column, i))
	}

	var column int
	err = huh.NewSelect[int]().Title("What is the column where the IPs are?").
		Options(options...).
		Value(&column).
		WithTheme(output.HuhTheme).
		Run()
	if err != nil {
		timber.Fatal(err, "failed to ask for column where IPs are")
	}
	return column
}

const (
	IpColumn                = "ip"
	NetworkAsnColumn        = "network.asn"
	NetworkIspColumn        = "network.isp"
	NetworkTypeColumn       = "network.type"
	LocationCityColumn      = "location.city"
	LocationStateColumn     = "location.state"
	LocationCountryColumn   = "location.country"
	LocationTimezoneColumn  = "location.timezone"
	LocationLongitudeColumn = "location.longitude"
	LocationLatitudeColumn  = "location.latitude"
	LocationGeoHashColumn   = "location.geohash"
	IPDataDeviceCountColumn = "ipdata.devicecount"
	IPDataBehaviorColumn    = "ipdata.behavior"
	IPDataCategoriesColumn  = "ipdata.categories"
	IPDataIPRiskColumn      = "ipdata.iprisk"
	IPDataEnrichedColumn    = "ipdata.enriched"
)

func SelectColumnsToAdd() []string {
	var columns []string
	options := []huh.Option[string]{
		huh.NewOption("IP", IpColumn),

		huh.NewOption("Network: Asn", NetworkAsnColumn),
		huh.NewOption("Network: ISP", NetworkIspColumn),
		huh.NewOption("Network: Type", NetworkTypeColumn),

		huh.NewOption("Location: City", LocationCityColumn),
		huh.NewOption("Location: State", LocationStateColumn),
		huh.NewOption("Location: Country", LocationCountryColumn),
		huh.NewOption("Location: Timezone", LocationTimezoneColumn),
		huh.NewOption("Location: Longitude", LocationLongitudeColumn),
		huh.NewOption("Location: Latitude", LocationLatitudeColumn),
		huh.NewOption("Location: Geo Hash", LocationGeoHashColumn),

		huh.NewOption("IP Data: Device Count", IPDataDeviceCountColumn),
		huh.NewOption("IP Data: Behavior", IPDataBehaviorColumn),
		huh.NewOption("IP Data: Categories", IPDataCategoriesColumn),
		huh.NewOption("IP Data: IP Risk", IPDataIPRiskColumn),
		huh.NewOption("IP Data: Enriched", IPDataEnrichedColumn),
	}
	for i, opt := range options {
		options[i] = opt.Selected(true)
	}
	err := huh.NewMultiSelect[string]().Title("Data you want to add").
		Description("These columns will get appended to your csv file.").
		Options(options...).
		WithTheme(output.HuhTheme).
		Run()
	if err != nil {
		timber.Fatal(err, "failed to run multi select")
	}
	return columns
}

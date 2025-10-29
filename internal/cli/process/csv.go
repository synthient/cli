package process

import (
	"encoding/csv"
	"slices"

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

type LookupColumnSelection struct {
	IP                bool
	NetworkAsn        bool
	NetworkIsp        bool
	NetworkType       bool
	LocationCity      bool
	LocationState     bool
	LocationCountry   bool
	LocationTimezone  bool
	LocationLongitude bool
	LocationLatitude  bool
	LocationGeoHash   bool
	IPDataDeviceCount bool
	IPDataBehavior    bool
	IPDataCategories  bool
	IPDataIPRisk      bool
	IPDataEnriched    bool
}

func SelectColumnsToAdd() LookupColumnSelection {
	var columns []string
	options := []huh.Option[string]{
		huh.NewOption("IP", "ip"),
		huh.NewOption("Network Asn", "network.asn"),
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
	return LookupColumnSelection{
		IP:         slices.Contains(columns, "ip"),
		NetworkAsn: slices.Contains(columns, "network.asn"),
	}
}

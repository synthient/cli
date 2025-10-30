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

type LookupColumnSelection int

const (
	IpColumn LookupColumnSelection = iota
	NetworkAsnColumn
	NetworkIspColumn
	NetworkTypeColumn
	LocationCityColumn
	LocationStateColumn
	LocationCountryColumn
	LocationTimezoneColumn
	LocationLongitudeColumn
	LocationLatitudeColumn
	LocationGeoHashColumn
	IPDataDeviceCountColumn
	IPDataBehaviorColumn
	IPDataCategoriesColumn
	IPDataIPRiskColumn
	IPDataEnrichedColumn
)

func SelectColumnsToAdd() []LookupColumnSelection {
	var columns []LookupColumnSelection
	options := []huh.Option[LookupColumnSelection]{
		huh.NewOption("IP", IpColumn),
		huh.NewOption("Network Asn", NetworkAsnColumn),
		huh.NewOption("Network ISP", NetworkIspColumn),
	}
	for i, opt := range options {
		options[i] = opt.Selected(true)
	}
	err := huh.NewMultiSelect[LookupColumnSelection]().Title("Data you want to add").
		Description("These columns will get appended to your csv file.").
		Options(options...).
		WithTheme(output.HuhTheme).
		Run()
	if err != nil {
		timber.Fatal(err, "failed to run multi select")
	}
	return columns
}

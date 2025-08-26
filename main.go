package main

import (
	"fmt"
	"net/http"
	"sync"
)

const (
	SPREADSHEET_ID    = "13goqcSbMKCq--j235GAN_GPO8ZH60cdd25xW1lhHLq0"
	SPREADSHEET_SHEET = "Chart Helper"
	START_ROW         = 2
	END_ROW           = 1500
)

type sheetValue struct {
	url       string // The Bureau of Labor Statistics URL which will generate the value in the valueCell
	valueCell string // The name of the cell where the generated value resides
	value     string // The value in the cell, as read from the spreadsheet
}

type inflationValue struct {
	url       string  // The URL which generated the value in value
	valueCell string  // The spreadsheet cell to which we want to write the value
	value     float64 // The value to write into the cell, if any
}

func getSheetValues(out chan sheetValue, client *http.Client, valueColumn, urlColumn string, wg *sync.WaitGroup) {
	defer func() {
		close(out)
		wg.Done()
	}()
	GetSheetValues(out, client, valueColumn, urlColumn)
}

func getInflationValues(in chan sheetValue, out chan inflationValue, cache *BLSCache, wg *sync.WaitGroup) {
	defer func() {
		close(out)
		wg.Done()
	}()
	for inval := range in {
		value, err := GetBlsValue(inval.url, cache)
		if err != nil {
			fmt.Println("No value from BLS for url: ", inval.url)
			continue
		}
		out <- inflationValue{
			url:       inval.url,
			valueCell: inval.valueCell,
			value:     value,
		}
	}
}

func updateSpreadsheetValues(in chan inflationValue, wg *sync.WaitGroup) {
	defer wg.Done()
	for inval := range in {
		fmt.Printf("Updating speadsheet, cell: %v, value: %v\n", inval.valueCell, inval.value)
		// TODO - actually update the spreadsheet
	}
}

func run(valueColumn string, urlColumn string) {
	// see: https://docs.google.com/spreadsheets/d/13goqcSbMKCq--j235GAN_GPO8ZH60cdd25xW1lhHLq0/edit?usp=sharing
	// Fetch the values from the spreadsheet
	// Fetch the inflation-corrected value from the CPI service
	// Update the spreadsheet
	client := Auth()
	cache := BLSCache{}
	sheetValChan := make(chan sheetValue)
	inflationValChan := make(chan inflationValue)
	var wg sync.WaitGroup
	wg.Add(3)
	go getSheetValues(sheetValChan, client, valueColumn, urlColumn, &wg)
	go getInflationValues(sheetValChan, inflationValChan, &cache, &wg)
	go updateSpreadsheetValues(inflationValChan, &wg)
	wg.Wait()
}

func main() {
	// Contractor equivalent salary data
	run("I", "J")

	// Salary data
	run("F", "G")
}

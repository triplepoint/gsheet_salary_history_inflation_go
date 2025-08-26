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
	valueCell string // The spreadsheet cell to which we want to write the value
}

type inflationValue struct {
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
			valueCell: inval.valueCell,
			value:     value,
		}
	}
}

func updateSpreadsheetValues(in chan inflationValue, client *http.Client, valueColumn string, wg *sync.WaitGroup) {
	defer wg.Done()
	UpdateSheetValues(in, client, valueColumn)
}

func run(valueColumn string, urlColumn string) {
	// see the (private) spreadsheet: https://docs.google.com/spreadsheets/d/13goqcSbMKCq--j235GAN_GPO8ZH60cdd25xW1lhHLq0/edit?usp=sharing
	// Relevant features are 2 columns, the "URL Column" and the "Value Column"
	// The URL column is prebuilt in the spreadsheet to create a URL call to the BLS inflation conversion site
	// The Value Column is where we want to write the value that we scrape out of the URL's result.
	//
	// - Fetch the values from the spreadsheet
	// - Fetch the inflation-corrected value from the CPI service
	// - Update the spreadsheet

	client := Auth()    // The Google spreadsheet client
	cache := NewCache() // A cache for avoiding calling the BLS when we already know the answer

	sheetValChan := make(chan sheetValue)
	inflationValChan := make(chan inflationValue)
	var wg sync.WaitGroup
	wg.Add(3)
	go getSheetValues(sheetValChan, client, valueColumn, urlColumn, &wg)
	go getInflationValues(sheetValChan, inflationValChan, cache, &wg)
	go updateSpreadsheetValues(inflationValChan, client, valueColumn, &wg)
	wg.Wait()
}

func main() {
	// Salary data
	run("F", "G")

	// Contractor equivalent salary data
	run("I", "J")
}

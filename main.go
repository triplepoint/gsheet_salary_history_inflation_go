/*
gsheet_salary_history_inflation_go populates inflation-corrected data in a spreadsheet.
see the (private) spreadsheet: https://docs.google.com/spreadsheets/d/13goqcSbMKCq--j235GAN_GPO8ZH60cdd25xW1lhHLq0/edit?usp=sharing

Generally, this program:
- Fetches a list of URLs from the spreadsheet
- Uses those URLs to fetch the inflation-corrected value from the Bureau of Labor Statistics' CPI service
- Writes those fetched values back to the spreadsheet
*/
package main

import (
	"fmt"
	"sync"
)

type sheetValue struct {
	url       string // The Bureau of Labor Statistics URL which will generate the value in the valueCell
	valueCell string // The spreadsheet cell to which we want to write the value
}

type inflationValue struct {
	valueCell string  // The spreadsheet cell to which we want to write the value
	value     float64 // The value to write into the cell, if any
}

func getInflationValues(in chan sheetValue, out chan inflationValue, cache *BLSCache) {
	defer close(out)
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

// Perform the ETL action for a pair of columns.
// The valueColumn is a letter-name column in the spreadsheet where the inflation-corrected value will be written.
// The urlColumn is a letter-name column in the spreadsheet where the prepared URL to call on the BLS site is read.
func run(valueColumn string, urlColumn string) {
	client := Auth()    // The Google spreadsheet client
	cache := NewCache() // A cache for avoiding calling the BLS when we already know the answer

	sheetValChan := make(chan sheetValue)
	inflationValChan := make(chan inflationValue)

	var wg sync.WaitGroup
	wg.Go(func() { GetSheetValues(sheetValChan, client, valueColumn, urlColumn) })
	wg.Go(func() { getInflationValues(sheetValChan, inflationValChan, cache) })
	wg.Go(func() { UpdateSheetValues(inflationValChan, client, valueColumn) })
	wg.Wait()
}

func main() {
	// Salary data
	run("F", "G")

	// Contractor equivalent salary data
	run("I", "J")
}

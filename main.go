//main2.go

package main

import (
	"fmt"
	"os"

	"github.com/gocarina/gocsv"
)

type Record struct {
	TIME string `csv:"TIME"`
	LAT  string `csv:"LAT"`
	LON string `csv:"LON"`
}

func main() {
	// Open the CSV file
	file, err := os.Open("1136-varie.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Read the CSV file into a slice of Record structs
	var records []Record
	if err := gocsv.UnmarshalFile(file, &records); err != nil {
		panic(err)
	}

	// Print the records
	for _, record := range records {
		fmt.Printf("TIME: %s, LAT: %s, LON: %s\n", record.TIME, record.LAT, record.LON)
	}
}

//main2.go

package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"image/color"

	"github.com/gocarina/gocsv"

	sm "github.com/flopp/go-staticmaps"
	"github.com/fogleman/gg"
	"github.com/golang/geo/s2"
)

type Record struct {
	TIME string `csv:"TIME"`
	LAT  string `csv:"LAT"`
	LON  string `csv:"LON"`
}

func StringToFloat64(data string) float64 {
	data = strings.Replace(data, ",", ".", -1)
	if s, err := strconv.ParseFloat(data, 64); err == nil {
		return s
	} else {
		fmt.Println(err.Error())
		return 0.0
	}
}

func IntToString(n int) string {
	return strconv.Itoa(n)
}

func main() {
	// Open the CSV file
	file, err := os.Open("1136-varie.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	ctx := sm.NewContext()
	ctx.SetSize(1920, 1080)

	// Read the CSV file into a slice of Record structs
	var records []Record
	if err := gocsv.UnmarshalFile(file, &records); err != nil {
		panic(err)
	}

	//Read the csv file and len
	v := len(records) - 1
	i := 0
	path := make([]s2.LatLng, 0, 2)

	// Print the records
	for _, record := range records {
		if i == 0 {
			first := sm.NewMarker(s2.LatLngFromDegrees(StringToFloat64(record.LAT), StringToFloat64(record.LON)), color.RGBA{255, 0, 0, 255}, 16.0)
			ctx.AddObject(first)
			path = append(path, first.Position)
			fmt.Printf("First Lan, Lon: %f,%f\n", StringToFloat64(record.LAT), StringToFloat64(record.LON))
		}
		if i == v {
			last := sm.NewMarker(s2.LatLngFromDegrees(StringToFloat64(record.LAT), StringToFloat64(record.LON)), color.RGBA{0, 0, 255, 255}, 16.0)
			ctx.AddObject(last)
			path = append(path, last.Position)
			//ctx.SetCenter(s2.LatLngFromDegrees(StringToFloat64(record.LAT), StringToFloat64(record.LON)))
			fmt.Printf("Last Lan, Lon: %f,%f\n", StringToFloat64(record.LAT), StringToFloat64(record.LON))
		}
		fmt.Printf("TIME: %s, LAT: %s, LON: %s\n", record.TIME, record.LAT, record.LON)
		// Add all others plots
		path = append(path, s2.LatLngFromDegrees(StringToFloat64(record.LAT), StringToFloat64(record.LON)))
		i = i + 1
	}
	fmt.Printf("Total path: %s\n", IntToString(len(path)))
	// Total path
	ctx.AddObject(sm.NewPath(path, color.RGBA{0, 255, 0, 255}, 4.0))

	img, err := ctx.Render()
	if err != nil {
		panic(err)
	}

	if err := gg.SavePNG("1136.png", img); err != nil {
		panic(err)
	}
}

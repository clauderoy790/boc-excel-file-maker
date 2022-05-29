package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	boc "github.com/clauderoy790/bank-of-canada-interests-rates"
	"github.com/xuri/excelize/v2"
)

const header = "Historique taux des obligations\nhttp://www.banqueducanada.ca/taux/taux-dinteret/obligations-canadiennes/\n** À partir du 20/04/2021,Taux 1 an = taux 2 ans\n"
const startDate = "2014-10-24"
const exportPath = "/Users/clauderoy/Desktop/test.xls"

var bank boc.BOCInterests
var f *excelize.File

func main() {
	var err error
	bank, err = boc.NewBOCInterests()
	if err != nil {
		panic(fmt.Errorf("error creating boc: %w", err))
	}

	exportPath :=  "/Users/clauderoy/Desktop/test.xls"
	f = excelize.NewFile()
	writeOECSheet(exportPath)
	// Save spreadsheet
	if err := f.SaveAs(exportPath); err != nil {
		panic(fmt.Errorf("failed to write file: %w", err))
	}
	// // Initialize astilectron
	// var a, _ = astilectron.New(log.New(os.Stderr, "", 0), astilectron.Options{
	// 	AppName:            "<your app name>",
	// 	AppIconDefaultPath: "<your .png icon>",  // If path is relative, it must be relative to the data directory
	// 	AppIconDarwinPath:  "<your .icns icon>", // Same here
	// 	BaseDirectoryPath:  "<where you want the provisioner to install the dependencies>",
	// 	VersionAstilectron: "<version of Astilectron to utilize such as `0.33.0`>",
	// 	VersionElectron:    "<version of Electron to utilize such as `4.0.1` | `6.1.2`>",
	// })
	// defer a.Close()

	// // Start astilectron
	// if err := a.Start(); err != nil {
	// 	panic(fmt.Errorf("fail to start astilectron: %w", err))
	// }

	// // Blocking pattern
	// a.Wait()
}

func writeOECSheet(path string) {
	// Create a new sheet.
	sheet := "OEC"
	f.SetActiveSheet(0)
	f.SetSheetName("Sheet1", sheet)

	header := getHeader()

	for i, str := range header {
		err := f.SetCellValue(sheet, fmt.Sprintf("A%d", (i+1)), str)
		if err != nil {
			panic(fmt.Errorf("error setting sheet vlaue: %w", err))
		}
	}
	writeLine(sheet, "5", []string{"Taux en date du:", "1 a 3 ans", "1 an", "2 ans", "3 ans", "4 ans", "5 ans"})

	date, err := boc.FormatDate(startDate)
	if err != nil {
		panic(fmt.Errorf("invalid date format: %w", err))
	}
	sDate, err := boc.FormatDate(date)
	if err != nil {
		panic(fmt.Errorf("invalid start date formmat: %s", startDate))
	}
	dt := parseToDate(sDate)
	currDate := time.Date(dt.year, time.Month(dt.month), dt.day, 0, 0, 0, 0, time.Local)
	now := time.Now()
	tomorrow := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	tomorrow.Add(time.Hour * 25)
	line := 6
	for {
		data, err := getRowData(fmt.Sprintf("%v-%v-%v", currDate.Year(), int(currDate.Month()), currDate.Day()))
		if err != nil {
			panic(fmt.Errorf("error building row: %w", err))
		}
		writeLine(sheet, strconv.Itoa(line), data)
		currDate = currDate.Add(24 * time.Hour)
		line++
		if currDate.After(now) {
			break
		}
	}

}

var letters = []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K"}

func writeLine(sheet, line string, vals []string) {
	for i, val := range vals {
		f.SetCellValue(sheet, fmt.Sprintf("%s%s", letters[i], line), val)
	}
}

func getRowData(date string) ([]string, error) {
	obs, err := bank.GetObservationForDate(date)
	if err != nil {
		return []string{colDateString(date), "n/a", "n/a", "n/a", "n/a", "n/a", "n/a"}, nil
	}
	three, err := strconv.ParseFloat(obs.Yield3Year.V, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid 3 year yield value: %s", obs.Yield3Year.V)
	}

	five, err := strconv.ParseFloat(obs.Yield5Year.V, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid 5 year yield value: %s", obs.Yield5Year.V)
	}

	avg := (three + five) / 2
	avg = math.Round(avg*100) / 100
	return []string{colDateString(date), formatFloat(obs.Average1To3Year.V), formatFloat(obs.Yield2Year.V), formatFloat(obs.Yield2Year.V), formatFloat(obs.Yield3Year.V), formatFloat(fmt.Sprintf("%f", avg)), formatFloat(obs.Yield5Year.V)}, nil
}

func formatFloat(val string) string {
	f, _ := strconv.ParseFloat(val, 64)
	f /= 100
	return fmt.Sprintf("%.4f", f)
}

func getHeader() []string {
	var headers []string
	headers = strings.Split(header, "\n")
	headers = append(headers, "\n")
	return headers
}

type date struct {
	year, month, day int
}

func colDateString(date string) string {
	d := parseToDate(date)
	return fmt.Sprintf("%d/%d/%d", d.month, d.day, d.year)
}

func parseToDate(d string) date {
	parts := strings.Split(d, "-")
	y, _ := strconv.Atoi(parts[0])
	m, _ := strconv.Atoi(parts[1])
	day, _ := strconv.Atoi(parts[2])
	return date{
		year:  y,
		month: m,
		day:   day,
	}
}

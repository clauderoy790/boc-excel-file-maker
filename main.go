package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	boc "github.com/clauderoy790/bank-of-canada-interests-rates"
	"github.com/clauderoy790/boc-excel-file-maker/common"
	"github.com/clauderoy790/boc-excel-file-maker/treasury"
	"github.com/xuri/excelize/v2"
)

const oecHeader = "Historique taux des obligations\nhttp://www.banqueducanada.ca/taux/taux-dinteret/obligations-canadiennes/\n** À partir du 20/04/2021,Taux 1 an = taux 2 ans\n"

var treasHeader = "Historique taux des obligations\n\nhttps://home.treasury.gov/resource-center/data-chart-center/interest-rates/TextView?type=daily_treasury_yield_curve&field_tdr_date_value_month=" + fmt.Sprintf("%04d%02d", time.Now().Year(), int(time.Now().Month())) + "\n"

const startDateOEC = "2014-10-24"

var startDateTreasury = time.Date(2015, 6, 19, 0, 0, 0, 0, time.Local)

const exportPath = "/Users/clauderoy/Desktop/test.xlsx"

var bank boc.BOCInterests
var f *excelize.File

func main() {
	var err error
	bank, err = boc.NewBOCInterests()
	if err != nil {
		panic(fmt.Errorf("error creating boc: %w", err))
	}

	f = excelize.NewFile()
	writeOECSheet()
	if err := writeUSTresory(); err != nil {
		panic(fmt.Errorf("error writing traesury: %w", err))
	}
	WriteWallStPrime()
	// Save spreadsheet
	if err := f.SaveAs(exportPath); err != nil {
		panic(fmt.Errorf("failed to write file: %w", err))
	}
}

func writeUSTresory() error {
	sheet := "US Tresory"
	f.NewSheet(sheet)
	f.SetActiveSheet(1)

	header := getHeader(treasHeader)
	for i, str := range header {
		err := f.SetCellValue(sheet, fmt.Sprintf("A%d", (i+1)), str)
		if err != nil {
			panic(fmt.Errorf("error setting sheet vlaue: %w", err))
		}
	}

	currDate := startDateTreasury
	now := time.Now()
	tomorrow := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	tomorrow.Add(time.Hour * 25)
	line := 6
	prevMonth, curMonth := -1, 0
	var data *treasury.Treasury
	for {
		if prevMonth != curMonth {
			var err error
			data, err = treasury.FetchData(currDate)
			if err != nil {
				return fmt.Errorf("error fetching treasury data for date %s: %w", dateString(currDate), err)
			}

		}

		rowData := getTreasRowData(currDate, data)
		if err := f.SetSheetRow(sheet, fmt.Sprintf("A%v", line), &rowData); err != nil {
			panic(err)
		}
		currDate = currDate.Add(24 * time.Hour)
		prevMonth = curMonth
		curMonth = int(currDate.Month())
		line++
		if currDate.After(now) {
			break
		}
	}
	return nil
}

func getTreasRowData(date time.Time, treas *treasury.Treasury) []interface{} {
	props, err := treas.GetPropsForDate(dateString(date))
	if err != nil {
		return []interface{}{colDateString(date), "n/a", "n/a", "n/a", "n/a", "n/a", "n/a", "n/a", "n/a", "n/a"}
	}
	return []interface{}{colDateString(date), formatFloat(props.Bc1Year.Content), formatFloat(props.Bc2Year.Content), formatFloat(props.Bc3Year.Content), formatFloat(props.Bc4Year.Content), formatFloat(props.Bc5Year.Content), formatFloat(props.Bc6Year.Content), formatFloat(props.Bc7Year.Content), formatFloat(props.Bc8Year.Content), formatFloat(props.Bc10Year.Content)}
}

func dateString(dt time.Time) string {
	return fmt.Sprintf("%04d-%02d-%02d", dt.Year(), int(dt.Month()), dt.Day())
}

func WriteWallStPrime() {
	f.NewSheet("Wall St Prime")
	f.SetActiveSheet(2)
}

func writeOECSheet() {
	sheet := "OEC"
	f.SetActiveSheet(0)
	f.SetSheetName("Sheet1", sheet)

	// header
	header := getHeader(oecHeader)
	for i, str := range header {
		err := f.SetCellValue(sheet, fmt.Sprintf("A%d", (i+1)), str)
		if err != nil {
			panic(fmt.Errorf("error setting sheet vlaue: %w", err))
		}
	}
	if err := f.SetSheetRow(sheet, "A5", &[]interface{}{"Taux en date du:", "1 a 3 ans", "1 an", "2 ans", "3 ans", "4 ans", "5 ans"}); err != nil {
		panic(err)
	}

	date, err := boc.FormatDate(startDateOEC)
	if err != nil {
		panic(fmt.Errorf("invalid date format: %w", err))
	}
	dt := parseToDate(date)
	currDate := time.Date(dt.year, time.Month(dt.month), dt.day, 0, 0, 0, 0, time.Local)
	now := time.Now()
	tomorrow := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	tomorrow.Add(time.Hour * 25)
	line := 6
	for {
		data, err := getOECRowData(currDate)
		if err != nil {
			panic(fmt.Errorf("error building row: %w", err))
		}
		if err := f.SetSheetRow(sheet, fmt.Sprintf("A%v", line), &data); err != nil {
			panic(err)
		}
		currDate = currDate.Add(24 * time.Hour)
		line++
		if currDate.After(now) {
			break
		}
	}
}

func getOECRowData(date time.Time) ([]interface{}, error) {

	obs, err := bank.GetObservationForDate(dateString(date))
	if err != nil {
		return []interface{}{colDateString(date), "n/a", "n/a", "n/a", "n/a", "n/a", "n/a"}, nil
	}
	three, err := strconv.ParseFloat(obs.Yield3Year.V, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid 3 year yield value: %s", obs.Yield3Year.V)
	}

	five, err := strconv.ParseFloat(obs.Yield5Year.V, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid 5 year yield value: %s", obs.Yield5Year.V)
	}

	avg := common.Average(three, five)
	return []interface{}{
		colDateString(date), formatFloat(obs.Average1To3Year.V), formatFloat(obs.Yield2Year.V),
		formatFloat(obs.Yield2Year.V), formatFloat(obs.Yield3Year.V), formatFloat(fmt.Sprintf("%f", avg)),
		formatFloat(obs.Yield5Year.V)}, nil
}

func getHeader(header string) []string {
	var headers []string
	headers = strings.Split(header, "\n")
	headers = append(headers, "\n")
	return headers
}

type date struct {
	year, month, day int
}

func colDateString(date time.Time) string {
	d := parseToDate(dateString(date))
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

func formatFloat(val string) string {
	f, _ := strconv.ParseFloat(val, 64)
	f /= 100
	return fmt.Sprintf("%.4f", f)
}

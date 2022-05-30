package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	boc "github.com/clauderoy790/bank-of-canada-interests-rates"
	"github.com/clauderoy790/boc-excel-file-maker/common"
	"github.com/clauderoy790/boc-excel-file-maker/treasury"
	"github.com/xuri/excelize/v2"
)

const wsjSheet = "Wall St Prime"

const oecHeader = "Historique taux des obligations\nhttp://www.banqueducanada.ca/taux/taux-dinteret/obligations-canadiennes/\n** À partir du 20/04/2021,Taux 1 an = taux 2 ans\n"

var treasHeader = "Historique taux des obligations\n\nhttps://home.treasury.gov/resource-center/data-chart-center/interest-rates/TextView?type=daily_treasury_yield_curve&field_tdr_date_value_month=" + fmt.Sprintf("%04d%02d", time.Now().Year(), int(time.Now().Month())) + "\n"

const startDateOEC = "2014-10-24"

var startDateTreasury = time.Date(2015, 6, 19, 0, 0, 0, 0, time.Local)

const filePath = "/Users/clauderoy/Desktop/test.xlsx"

var bank boc.BOCInterests

func main() {
	var err error
	bank, err = boc.NewBOCInterests()
	if err != nil {
		panic(fmt.Errorf("error creating boc: %w", err))
	}

	f := excelize.NewFile()
	writeOECSheet(f)
	if err := writeUSTresory(f); err != nil {
		panic(fmt.Errorf("error writing traesury: %w", err))
	}
	if err := WriteWallStPrime(f); err != nil {
		panic(fmt.Errorf("error writing WSJ: %w", err))
	}
	f.SetActiveSheet(0)
	// Save spreadsheet
	if err := f.SaveAs(filePath); err != nil {
		panic(fmt.Errorf("failed to write file: %w", err))
	}
	f.Close()

	time.Sleep(time.Second * 2)

	if err := updateWSJ(); err != nil {
		panic(fmt.Errorf("error updating wsj: %w", err))
	}
}

func updateWSJ() error {
	var err error
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	_, err = getwsjData(f, "9")
	if err != nil {
		return fmt.Errorf("error getting wsj data: %w", err)
	}
	currentData, err := getwsjData(f, "10")
	if err != nil {
		return fmt.Errorf("error getting wsj data: %w", err)
	}
	newData, err := getwsjData(f, "11")
	if err != nil {
		return fmt.Errorf("error getting wsj data: %w", err)
	}
	if newData.wsj.val != currentData.wsj.val {
		if err := shiftData(f, "A"); err != nil {
			return fmt.Errorf("error shifting data: %w", err)
		}
	}
	if newData.can.val != currentData.can.val {
		if err := shiftData(f, "G"); err != nil {
			return fmt.Errorf("error shifting data: %w", err)
		}
	}
	if newData.us.val != currentData.can.val {
		if err := shiftData(f, "J"); err != nil {
			return fmt.Errorf("error shifting data: %w", err)
		}
	}
	err = f.RemoveRow(wsjSheet, 11)
	if err != nil {
		return err
	}
	if err := f.SaveAs(filePath); err != nil {
		panic(fmt.Errorf("failed to write file: %w", err))
	}
	return nil
}

func shiftData(f *excelize.File, col string) error {
	c := []rune(col)[0]
	for i := 0; i < 2; i++ {
		for row := 6; row <= 11; row++ {
			currCol := string(rune(int(c) + i))
			curr := currCol + fmt.Sprintf("%d", row)
			previous := currCol + fmt.Sprintf("%d", row-1)
			v, err := f.GetCellValue(wsjSheet, curr)
			fmt.Println("Current has: ", v)
			fmt.Println("presvious is: ", previous)
			if err != nil {
				return err
			}

			err = f.SetCellValue(wsjSheet, previous, nil)
			time.Sleep(time.Second / 100)
			err = f.SetCellValue(wsjSheet, previous, v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func getwsjData(f *excelize.File, line string) (*wsjData, error) {
	rate, err := newRate(f, line, "A")
	if err != nil {
		return nil, err
	}
	us, err := newRate(f, line, "G")
	if err != nil {
		return nil, err
	}
	can, err := newRate(f, line, "J")
	if err != nil {
		return nil, err
	}
	return &wsjData{
		wsj: *rate,
		us:  *us,
		can: *can,
	}, nil

}

type wsjData struct {
	wsj rate
	us  rate
	can rate
}

type rate struct {
	date time.Time
	val  string
}

func newRate(f *excelize.File, line, col string) (*rate, error) {
	dt, err := f.GetCellValue(wsjSheet, col+line)
	if err != nil {
		return nil, err
	}
	next := string(rune(int([]rune(col)[0]) + 1))
	perc, err := f.GetCellValue(wsjSheet, next+line)
	if err != nil {
		return nil, err
	}

	return &rate{
		date: toDate(dt),
		val:  perc,
	}, nil
}

func toDate(dt string) time.Time {
	parts := strings.Split(dt, "-")
	d, _ := strconv.Atoi(parts[0])

	month := time.January
	for {
		if strings.HasPrefix(month.String(), parts[1]) {
			break
		}
		month = time.Month(int(month) + 1)
	}

	y, _ := strconv.Atoi("20" + parts[2])
	return time.Date(y, month, d, 0, 0, 0, 0, time.Local)
}

func writeUSTresory(f *excelize.File) error {
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

func WriteWallStPrime(f *excelize.File) error {
	sheet := wsjSheet
	f.NewSheet(sheet)
	f.SetActiveSheet(2)

	//firt line
	_ = f.SetCellValue(sheet, "A1", "Wall Street #45")
	_ = f.SetCellValue(sheet, "B1", "https://www.wsj.com/market-data/bonds")
	_ = f.SetCellValue(sheet, "G1", "Prime US BNC(#3)")
	_ = f.SetCellValue(sheet, "J1", "Prime CAN BNC (#2)")
	_ = f.SetCellValue(sheet, "G2", "https://www.bnc.ca/fr/taux-et-analyses/taux-dinteret-et-rendements/taux-de-base.html")

	//existing
	writeFirst2Cells(f, sheet, "5", "19-Sep-19", "5.00%")
	writeFirst2Cells(f, sheet, "6", "31-Oct-19", "4.75%")
	writeFirst2Cells(f, sheet, "7", "4-Mar-20", "4.25%")
	writeFirst2Cells(f, sheet, "8", "16-Mar-20", "3.25%")
	writeFirst2Cells(f, sheet, "9", "17-Mar-22", "3.50%")
	writeFirst2Cells(f, sheet, "10", "4-May-22", "4.00%")

	writeBNCells(f, sheet, "5", "20-Sep-19", float64(5.5), "25-Oct-18", float64(3.95))
	writeBNCells(f, sheet, "6", "1-Nov-19", float64(5.25), "6-Mar-20", float64(3.45))
	writeBNCells(f, sheet, "7", "6-Mar-20", float64(4.75), "17-Mar-20", float64(2.95))
	writeBNCells(f, sheet, "8", "17-Mar-20", float64(3.75), "31-Mar-20", float64(2.45))
	writeBNCells(f, sheet, "9", "17-Mar-22", float64(4.00), "3-Mar-22", float64(2.70))
	writeBNCells(f, sheet, "10", "5-May-22", float64(4.50), "14-Mar-22", float64(3.20))

	us, can, err := getBNData()
	if err != nil {
		return fmt.Errorf("error getting BN data: %w", err)
	}
	val, err := getFedData()
	if err != nil {
		return fmt.Errorf("error getting WSJ data: %w", err)
	}

	now := time.Now()
	writeBNCells(f, sheet, "11", wsjDate(now), us, wsjDate(now), can)
	writeFirst2Cells(f, sheet, "11", wsjDate(now), percent(val))
	return nil
}

func wsjDate(date time.Time) string {
	return fmt.Sprintf("%d-%s-%s", date.Day(), date.Month().String()[:3], strconv.Itoa(date.Year())[2:])
}

func writeBNCells(f *excelize.File, sheet, line, v1 string, us float64, v3 string, can float64) {
	u := percent(us)
	c := percent(can)
	_ = f.SetCellValue(sheet, "G"+line, v1)
	_ = f.SetCellValue(sheet, "H"+line, u)
	_ = f.SetCellValue(sheet, "J"+line, v3)
	_ = f.SetCellValue(sheet, "K"+line, c)
}

func percent(us float64) string {
	return fmt.Sprintf("%.2f", us) + "%"
}

func writeFirst2Cells(f *excelize.File, sheet, line, v1, v2 string) {
	_ = f.SetCellValue(sheet, "A"+line, v1)
	_ = f.SetCellValue(sheet, "B"+line, v2)
}

func writeOECSheet(f *excelize.File) {
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

func getBNData() (us, can float64, err error) {
	us, can = 0, 0
	path := "https://www.bnc.ca/fr/taux-et-analyses/taux-dinteret-et-rendements/taux-de-base.html"
	var resp *http.Response
	if resp, err = http.Get(path); err != nil {
		err = fmt.Errorf("error making request: %w", err)
		return
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		err = fmt.Errorf("error reading body: %w", err)
		return
	}
	bodyStr := string(bodyBytes)
	if resp.StatusCode != 200 {
		err = fmt.Errorf("invalid status code: %d\nbody: %s", resp.StatusCode, bodyStr)
		return
	}
	var document *goquery.Document
	document, err = goquery.NewDocumentFromReader(strings.NewReader(bodyStr))
	if err != nil {
		err = fmt.Errorf("failed to create document: %w", err)
		return

	}
	sel := document.Find(".nbc-table tbody")
	if sel != nil {
		sel.Find("tr").Each(func(i int, s *goquery.Selection) {
			if i == 0 {
				return
			}
			text := strings.TrimSpace(s.Text())
			isUS := true
			if i == 1 && strings.Contains(text, "CA") {
				isUS = false
			}
			fl := float64(0)
			for i, r := range text {
				if unicode.IsDigit(r) {
					text = text[i:]
					break
				}
			}
			fl, err = strconv.ParseFloat(text, 64)
			if err != nil {
				err = fmt.Errorf("invalid rate: %s \n err: %w", text, err)
				return
			}
			if isUS {
				us = fl
			} else {
				can = fl
			}
		})
	}

	if us == 0 || can == 0 {
		err = fmt.Errorf("failed to find all rates US: %v, CAN: %v", us, can)
	}
	return
}

func getFedData() (fl float64, err error) {
	path := "http://www.fedprimerate.com/wall_street_journal_prime_rate_history.htm"
	var resp *http.Response
	if resp, err = http.Get(path); err != nil {
		return 0, fmt.Errorf("error making request: %w", err)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return 0, fmt.Errorf("error reading body: %w", err)
	}
	bodyStr := string(bodyBytes)
	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("invalid status code: %d\nbody: %s", resp.StatusCode, bodyStr)
	}
	var document *goquery.Document
	document, err = goquery.NewDocumentFromReader(strings.NewReader(bodyStr))
	if err != nil {
		return 0, fmt.Errorf("failed to create document: %w", err)
	}
	document.Find("tr").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		currentText := "(The Current U.S. Prime Rate)"
		if !strings.Contains(text, currentText) {
			return
		}
		s.Find("td").Each(func(j int, s *goquery.Selection) {
			if j != 1 {
				return
			}
			text := s.Text()
			text = strings.TrimSpace(strings.ReplaceAll(text, currentText, ""))
			fl, err = strconv.ParseFloat(text, 64)
			if err != nil {
				err = fmt.Errorf("invalid wsj rate: %s \n err: %w", text, err)
				return
			}
		})
	})
	return fl, nil
}

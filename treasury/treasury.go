package treasury

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	xj "github.com/basgys/goxml2json"
	"github.com/clauderoy790/boc-excel-file-maker/common"
)

const downloadPath = "https://home.treasury.gov/resource-center/data-chart-center/interest-rates/pages/xml?data=daily_treasury_yield_curve&field_tdr_date_value_month="
const cachePath = "./cache"

func newTreasury(d time.Time) *Treasury {
	path := fmt.Sprintf("%s%02d%02d", downloadPath, d.Year(), int(d.Month()))
	return &Treasury{
		path:  path,
		props: make(map[string]*Properties),
	}
}

type Treasury struct {
	path  string
	data  *TreasuryData
	props map[string]*Properties
}

func (t *Treasury) fetchData() ([]byte, error) {
	var resp *http.Response
	var err error
	if resp, err = http.Get(t.path); err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body: %w", err)
	}
	bodyStr := string(bytes)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got error status code: %d with response: %s", resp.StatusCode, bodyStr)
	}
	jsonData, err := xj.Convert(strings.NewReader(bodyStr))
	if err != nil {
		return nil, fmt.Errorf("fail to convert XML to json: %w", err)
	}
	return jsonData.Bytes(), nil

}

func (t *Treasury) setDataFromBytes(jsonBytes []byte) error {
	treasury := new(TreasuryData)
	if err := json.Unmarshal(jsonBytes, treasury); err != nil {
		return fmt.Errorf("error while unmarshalling: %w", err)
	}

	// remove unless text in date
	t.data = treasury
	for _, entry := range treasury.Feed.Entry {
		prop := entry.Content.Properties
		dateSuffix := "T00:00:00"
		if !strings.HasSuffix(prop.Date.Content, dateSuffix) {
			return fmt.Errorf("invalid date format: %s", prop.Date.Content)
		}
		date := strings.ReplaceAll(prop.Date.Content, dateSuffix, "")
		prop.Date.Content = date

		//custom vals
		prop.Bc4Year.Content = t.getAverageValue(prop.Bc3Year, prop.Bc5Year)
		prop.Bc6Year.Content = t.getAverageValue(prop.Bc5Year, prop.Bc7Year)
		prop.Bc8Year.Content = t.getAverageValue(prop.Bc7Year, prop.Bc10Year)

		t.props[date] = prop
	}
	return nil
}

func (t *Treasury) getAverageValue(v1, v2 V) string {
	f1, err := strconv.ParseFloat(v1.Content, 64)
	if err != nil {
		panic(err)
	}
	f2, err := strconv.ParseFloat(v2.Content, 64)
	if err != nil {
		panic(err)
	}
	f := common.Average(f1, f2)

	return fmt.Sprintf("%.2f", f)
}

func (t *Treasury) GetPropsForDate(date string) (*Properties, error) {
	props := t.props[date]
	if props == nil {
		return nil, fmt.Errorf("no properties for date %s", date)
	}
	return props, nil
}

func FetchData(dt time.Time) (*Treasury, error) {
	t := newTreasury(dt)
	ex, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("unable to get executable path: %w", err)
	}
	wd := filepath.Dir(ex)
	cache := path.Join(wd, cachePath)
	if _, err := os.Stat(cache); os.IsNotExist(err) {
		os.Mkdir(cache, 0755)
	}
	jsonFile := path.Join(cache, dateString(dt)+".json")
	var data []byte
	if _, err := os.Stat(jsonFile); err == nil {
		data, err = ioutil.ReadFile(jsonFile)
		if err != nil {
			return nil, fmt.Errorf("error restoring cache file %s: %w", jsonFile, err)
		}
	} else {
		data, err = t.fetchData()
		if err != nil {
			return nil, fmt.Errorf("error fetching data: %w", err)
		}
		if err := ioutil.WriteFile(jsonFile, data, 0755); err != nil {
			return nil, fmt.Errorf("error writing cached file %s: %w", jsonFile, err)
		}
	}
	t.setDataFromBytes(data)
	return t, nil
}

type TreasuryData struct {
	Feed Feed `json:"feed"`
}

type V struct {
	Content string `json:"#content"`
	Type    string `json:"-type"`
}

type Properties struct {
	Date            V `json:"NEW_DATE"`
	Bc2Year         V `json:"BC_2YEAR"`
	Bc3Month        V `json:"BC_3MONTH"`
	Bc30YearDisplay V `json:"BC_30YEARDISPLAY"`
	Bc2Month        V `json:"BC_2MONTH"`
	Bc5Year         V `json:"BC_5YEAR"`
	Bc3Year         V `json:"BC_3YEAR"`
	Bc7Year         V `json:"BC_7YEAR"`
	Bc10Year        V `json:"BC_10YEAR"`
	Bc6Month        V `json:"BC_6MONTH"`
	Bc1Year         V `json:"BC_1YEAR"`
	Bc20Year        V `json:"BC_20YEAR"`
	Bc30Year        V `json:"BC_30YEAR"`
	Bc1Month        V `json:"BC_1MONTH"`
	Bc4Year         V // custom
	Bc6Year         V // custom
	Bc8Year         V // custom
}
type Content struct {
	Type       string      `json:"-type"`
	Properties *Properties `json:"properties"`
}
type Entry struct {
	Content *Content `json:"content"`
}
type Feed struct {
	Updated time.Time `json:"updated"`
	Entry   []*Entry  `json:"entry"`
}

func dateString(dt time.Time) string {
	return fmt.Sprintf("%04d-%02d-%02d", dt.Year(), int(dt.Month()), dt.Day())
}

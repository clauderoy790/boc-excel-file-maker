package treasury

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	xj "github.com/basgys/goxml2json"
)

const downloadPath = "https://home.treasury.gov/resource-center/data-chart-center/interest-rates/pages/xml?data=daily_treasury_yield_curve&field_tdr_date_value_month="

func newTreasury(d time.Time) *Treasury {
	path := fmt.Sprintf("%s%02d%02d", downloadPath, d.Year(), int(d.Month()))
	return &Treasury{
		path: path,
	}
}

type Treasury struct {
	path string
	data *TreasuryData
}

func (t *Treasury) fetchData() (*TreasuryData, error) {
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
	treasury := new(TreasuryData)
	if err := json.Unmarshal(jsonData.Bytes(), treasury); err != nil {
		return nil, fmt.Errorf("error while unmarshalling: %w", err)
	}

	// remove unless text in date
	t.data = treasury
	for _, entry := range treasury.Feed.Entry {
		prop := entry.Content.Properties
		dateSuffix := "T00:00:00"
		if !strings.HasSuffix(prop.Date.Content, dateSuffix) {
			return nil, fmt.Errorf("invalid date format: %s", prop.Date.Content)
		}
		prop.Date.Content = strings.ReplaceAll(prop.Date.Content, dateSuffix, "")
	}
	return t.data, nil
}

func (t *Treasury) GetDataForDate(date string)  {
	t.data.

}

func FetchData() (*TreasuryData, error) {
	t := newTreasury(time.Now())
	return t.fetchData()
}

type TreasuryData struct {
	Feed Feed `json:"feed"`
}

type Title struct {
	Type string `json:"-type"`
}
type Author struct {
	Name string `json:"name"`
}
type Link struct {
	Rel   string `json:"-rel"`
	Title string `json:"-title"`
	Href  string `json:"-href"`
}
type Category struct {
	Term   string `json:"-term"`
	Scheme string `json:"-scheme"`
}
type V struct {
	Content string `json:"#content"`
	Type    string `json:"-type"`
}

type Properties struct {
	Bc2Year         V `json:"BC_2YEAR"`
	Date            V `json:"NEW_DATE"`
	Bc3Month        V `json:"BC_3MONTH"`
	Bc30Yeardisplay V `json:"BC_30YEARDISPLAY"`
	Bc2Month        V `json:"BC_2MONTH"`
	Bc5Year         V `json:"BC_5YEAR"`
	Bc3Year         V `json:"BC_3YEAR"`
	Bc7Year         V `json:"BC_7YEAR"`
	Bc10Year        V `json:"BC_10YEAR"`
	Bc6Month        V `json:"BC_6MONTH"`
	Bc1Year         V `json:"BC_1YEAR"`
	Bc20Year        V `json:"BC_20YEAR"`
	Bc30Year        V `json:"BC_30YEAR"`
	ID              V `json:"Id"`
	Bc1Month        V `json:"BC_1MONTH"`
}
type Content struct {
	Type       string      `json:"-type"`
	Properties *Properties `json:"properties"`
}
type Entry struct {
	ID       string    `json:"id"`
	Title    Title     `json:"title"`
	Updated  time.Time `json:"updated"`
	Author   *Author   `json:"author"`
	Link     *Link     `json:"link"`
	Category *Category `json:"category"`
	Content  *Content  `json:"content"`
}
type Feed struct {
	Base    string    `json:"-base"`
	D       string    `json:"-d"`
	M       string    `json:"-m"`
	Title   *Title    `json:"title"`
	ID      string    `json:"id"`
	Updated time.Time `json:"updated"`
	Entry   []*Entry  `json:"entry"`
	Xmlns   string    `json:"-xmlns"`
	Link    *Link     `json:"link"`
}

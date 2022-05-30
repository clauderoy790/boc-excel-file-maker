package treasury

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_NewTreasury(t *testing.T) {
	a := assert.New(t)
	now := time.Now()
	treas := newTreasury(now)
	expectedPath := downloadPath + fmt.Sprintf("%04d%02d", now.Year(), int(now.Month()))
	a.Equal(expectedPath, treas.path)

}

func Test_FetchData(t *testing.T) {
	a := assert.New(t)
	d := time.Date(2022, 2, 1, 0, 0, 0, 0, time.Local)
	treas := newTreasury(d)
	err := treas.fetchData()
	a.NoError(err)
	data := treas.data

	for _, d := range data.Feed.Entry {
		props := d.Content.Properties
		a.False(strings.HasSuffix(props.Date.Content, "T00:00:00"))
	}
	a.Equal(len(data.Feed.Entry), len(treas.props))
	p, err := treas.GetPropsForDate("2022-02-01")
	a.NotNil(p)
	a.NoError(err)
	a.Equal("0.19", p.Bc3Month.Content)
	a.Equal("0.78", p.Bc1Year.Content)
	a.Equal("1.18", p.Bc2Year.Content)
	a.Equal("1.39", p.Bc3Year.Content)
	a.Equal("1.63", p.Bc5Year.Content)
	a.Equal("1.76", p.Bc7Year.Content)
	a.Equal("1.81", p.Bc10Year.Content)
	a.NotEmpty(p.Bc4Year.Content)
	a.NotEmpty(p.Bc6Year.Content)
	a.NotEmpty(p.Bc8Year.Content)


	p, err = treas.GetPropsForDate("2022-02-16")
	a.NotNil(p)
	a.NoError(err)
	a.Equal("0.38", p.Bc3Month.Content)
	a.Equal("1.09", p.Bc1Year.Content)
	a.Equal("1.52", p.Bc2Year.Content)
	a.Equal("1.75", p.Bc3Year.Content)
	a.Equal("1.90", p.Bc5Year.Content)
	a.Equal("2.00", p.Bc7Year.Content)
	a.Equal("2.03", p.Bc10Year.Content)
	a.NotEmpty(p.Bc4Year.Content)
	a.NotEmpty(p.Bc6Year.Content)
	a.NotEmpty(p.Bc8Year.Content)

	p, err = treas.GetPropsForDate("2022-02-25")
	a.NotNil(p)
	a.NoError(err)
	a.Equal("0.33", p.Bc3Month.Content)
	a.Equal("1.13", p.Bc1Year.Content)
	a.Equal("1.55", p.Bc2Year.Content)
	a.Equal("1.76", p.Bc3Year.Content)
	a.Equal("1.86", p.Bc5Year.Content)
	a.Equal("1.96", p.Bc7Year.Content)
	a.Equal("1.97", p.Bc10Year.Content)
	a.NotEmpty(p.Bc4Year.Content)
	a.NotEmpty(p.Bc6Year.Content)
	a.NotEmpty(p.Bc8Year.Content)

}

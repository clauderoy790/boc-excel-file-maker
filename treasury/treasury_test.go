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
	data, err := treas.fetchData()
	a.NoError(err)
	a.NotNil(data)

	for _, d := range data.Feed.Entry {
		props := d.Content.Properties
		a.False(strings.HasSuffix(props.Date.Content, "T00:00:00"))
	}
}

package main

import (
	"reflect"
	"testing"
	"time"
)

func Test_getBNData(t *testing.T) {
	tests := []struct {
		name    string
		wantUs  float64
		wantCan float64
		wantErr bool
	}{
		{
			name:    "success",
			wantUs:  4.5,
			wantCan: 3.2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUs, gotCan, err := getBNData()
			if (err != nil) != tt.wantErr {
				t.Errorf("getBNData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotUs != tt.wantUs {
				t.Errorf("getBNData() gotUs = %v, want %v", gotUs, tt.wantUs)
			}
			if gotCan != tt.wantCan {
				t.Errorf("getBNData() gotCan = %v, want %v", gotCan, tt.wantCan)
			}
		})
	}
}

func Test_getFedData(t *testing.T) {
	tests := []struct {
		name    string
		wantFl  float64
		wantErr bool
	}{
		{
			name:   "success",
			wantFl: 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFl, err := getFedData()
			if (err != nil) != tt.wantErr {
				t.Errorf("getFedData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotFl != tt.wantFl {
				t.Errorf("getFedData() = %v, want %v", gotFl, tt.wantFl)
			}
		})
	}
}

func Test_wsjDate(t *testing.T) {
	tests := []struct {
		name string
		date time.Time
		want string
	}{
		{
			name: "success",
			date: time.Now(),
			want: "30-May-22",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := wsjDate(tt.date); got != tt.want {
				t.Errorf("wsjDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_percent(t *testing.T) {
	tests := []struct {
		name string
		us   float64
		want string
	}{
		{
			name: "success",
			us:   3.45565,
			want: "3.45%",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := percent(tt.us); got != tt.want {
				t.Errorf("percent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_toDate(t *testing.T) {
	tests := []struct {
		name string
		dt   string
		want time.Time
	}{
		{
			name: "success",
			dt:   "4-May-22",
			want: time.Date(2022, time.May, 4, 0, 0, 0, 0, time.Local),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toDate(tt.dt); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

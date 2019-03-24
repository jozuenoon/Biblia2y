package messenger

import (
	"reflect"
	"testing"
	"time"
)

func Test_parseSetTimeCommand(t *testing.T) {
	loc, err := time.LoadLocation("UTC")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		msg     string
		want    time.Time
		wantErr bool
	}{
		{"basic", "14:00", time.Date(0, 1, 1, 14, 0, 0, 0, loc), false},
		{"basic spaces", " 14:00 ", time.Date(0, 1, 1, 14, 0, 0, 0, loc), false},
		{"basic spaces", "14:00 ", time.Date(0, 1, 1, 14, 0, 0, 0, loc), false},
		{"basic spaces", " 14:00", time.Date(0, 1, 1, 14, 0, 0, 0, loc), false},
		{"basic spaces", "set time 14:00", time.Date(0, 1, 1, 14, 0, 0, 0, loc), false},
		{"basic spaces", "set time 14:00 ", time.Date(0, 1, 1, 14, 0, 0, 0, loc), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSetTimeCommand(tt.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSetTimeCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseSetTimeCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

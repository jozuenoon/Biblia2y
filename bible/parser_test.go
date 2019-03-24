package bible

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/go-kit/kit/log"
)

func TestParser_Parse(t *testing.T) {
	logger := log.NewNopLogger()

	s, err := New("", "", "", logger)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		buf     io.Reader
		want    *Verse
		wantErr bool
	}{
		{
			"single",
			strings.NewReader("1kor 1,1"),
			&Verse{start: 28701, end: 0},
			false,
		},
		{
			"single1",
			strings.NewReader("1kor 2,1"),
			&Verse{start: 28732, end: 0},
			false,
		},
		{
			"range",
			strings.NewReader("1kor 1,1-31"),
			&Verse{start: 28701, end: 28731},
			false,
		},
		{
			"chapter_range",
			strings.NewReader("1kor 1"),
			&Verse{start: 28701, end: 28731},
			false,
		},
		{
			"range1",
			strings.NewReader("1kor 1,1-2"),
			&Verse{start: 28701, end: 28702},
			false,
		},
		{
			"range2",
			strings.NewReader("1kor 1,1-2,1"),
			&Verse{start: 28701, end: 28732},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.buf, s)
			got, err := p.Parse()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parser.Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

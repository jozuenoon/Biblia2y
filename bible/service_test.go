package bible

import (
	"testing"
)

func Test_service_getIndexFromChapterLabel(t *testing.T) {
	tests := []struct {
		name    string
		idxMap  map[Label]int
		label   Label
		want    int
		wantErr bool
	}{
		{
			"basic",
			map[Label]int{
				Label("001001001"): 1,
				Label("002001001"): 2,
				Label("003001001"): 3,
				Label("004001001"): 4,
				Label("005001001"): 5,
			},
			Label("001001"),
			1,
			false,
		},
		{
			"basic",
			map[Label]int{
				Label("001001001"): 1,
				Label("002001001"): 2,
				Label("003001001"): 3,
				Label("004001001"): 4,
				Label("005001001"): 5,
			},
			Label("002001"),
			2,
			false,
		},
		{
			"basic",
			map[Label]int{
				Label("001001001"): 1,
				Label("002001001"): 2,
				Label("003001001"): 3,
				Label("004001001"): 4,
				Label("005001001"): 5,
			},
			Label("003001"),
			3,
			false,
		},
		{
			"basic",
			map[Label]int{
				Label("001001001"): 1,
				Label("002001001"): 2,
				Label("003001001"): 3,
				Label("004001001"): 4,
				Label("005001001"): 5,
			},
			Label("004001"),
			4,
			false,
		},
		{
			"basic",
			map[Label]int{
				Label("001001001"): 1,
				Label("002001001"): 2,
				Label("003001001"): 3,
				Label("004001001"): 4,
				Label("005001001"): 5,
			},
			Label("005001"),
			5,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				idxMap: tt.idxMap,
			}
			got, err := s.getIndexFromChapterLabel(tt.label)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.getIndexFromChapterLabel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("service.getIndexFromChapterLabel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_service_GetChapterStartIndex(t *testing.T) {

	tests := []struct {
		name     string
		labelMap map[int]Label
		index    int
		want     int
		wantErr  bool
	}{
		{
			"basic",
			map[int]Label{
				1: Label("001001001"),
				2: Label("001001002"),
				3: Label("001002001"),
				4: Label("001002002"),
			},
			4,
			3,
			false,
		},
		{
			"basic1",
			map[int]Label{
				0: Label("001000000"),
				1: Label("001001001"),
				2: Label("001001002"),
				3: Label("001001003"),
				4: Label("001001004"),
			},
			4,
			1,
			false,
		},
		{
			"error",
			map[int]Label{},
			4,
			0,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				labelMap: tt.labelMap,
			}
			got, err := s.GetChapterStartIndex(tt.index)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.GetChapterStartIndex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("service.GetChapterStartIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}
